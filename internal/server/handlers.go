package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/fserg/md-to-html/internal/converter"
	"github.com/fserg/md-to-html/internal/version"
	"github.com/go-chi/chi/v5"
)

const defaultDocumentTitle = "Document"

type Server struct {
	cfg   Config
	conv  *converter.Converter
	store *PreviewStore
	log   *slog.Logger
}

type convertRequest struct {
	Markdown string `json:"markdown"`
	Title    string `json:"title,omitempty"`
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if !hasJSONContentType(r.Header.Get("Content-Type")) {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{
			"detail": "content-type must be application/json",
		})
		return
	}

	var payload convertRequest
	if err := decodeJSON(r, &payload); err != nil {
		s.writeDecodeError(w, err)
		return
	}

	result, err := s.convertMarkdown(payload.Markdown, payload.Title)
	if err != nil {
		s.writeConvertError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.HTML)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": version.Version})
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"template_loaded": s.conv != nil,
	})
}

func (s *Server) handleHome(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("UI coming in phase 4"))
}

func (s *Server) handleUIConvert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.writeDecodeError(w, err)
		return
	}

	result, err := s.convertMarkdown(r.Form.Get("markdown"), r.Form.Get("title"))
	if err != nil {
		s.writeConvertError(w, err)
		return
	}

	filename := htmlFilename(result.Title)
	previewID := s.store.Put(result.HTML, "text/html; charset=utf-8", filename)
	downloadID := s.store.Put(result.HTML, "text/html; charset=utf-8", filename)

	fragment := fmt.Sprintf(
		`<div><p>Result ready</p><a href="/preview/%s" target="_blank" rel="noopener">Preview</a> <a href="/download/%s">Download</a></div>`,
		previewID,
		downloadID,
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fragment))
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, ok := s.store.Take(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", contentTypeOrDefault(item.mime))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(item.html)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, ok := s.store.Take(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", contentTypeOrDefault(item.mime))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": item.filename,
	}))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(item.html)
}

func (s *Server) convertMarkdown(markdown, title string) (converter.Result, error) {
	if strings.TrimSpace(markdown) == "" {
		return converter.Result{}, errEmptyMarkdown
	}

	if int64(len([]byte(markdown))) > s.cfg.MaxMarkdownBytes {
		return converter.Result{}, errMarkdownTooLarge{limit: s.cfg.MaxMarkdownBytes}
	}

	fallbackTitle := strings.TrimSpace(title)
	if fallbackTitle == "" {
		fallbackTitle = defaultDocumentTitle
	}

	result, err := s.conv.Convert([]byte(markdown), fallbackTitle)
	if err != nil {
		return converter.Result{}, fmt.Errorf("convert markdown: %w", err)
	}

	return result, nil
}

func (s *Server) writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
			"detail": fmt.Sprintf("request exceeds %d bytes", s.cfg.MaxRequestBytes),
		})
		return
	}

	writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "invalid request payload"})
}

func (s *Server) writeConvertError(w http.ResponseWriter, err error) {
	var markdownTooLarge errMarkdownTooLarge
	switch {
	case errors.Is(err, errEmptyMarkdown):
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
	case errors.As(err, &markdownTooLarge):
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
			"detail": markdownTooLarge.Error(),
		})
	default:
		s.log.Error("convert_failed", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"detail": err.Error()})
	}
}

func hasJSONContentType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	return err == nil && mediaType == "application/json"
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if len(extra) > 0 {
		return errors.New("unexpected trailing JSON data")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

func htmlFilename(title string) string {
	name := strings.TrimSpace(title)
	if name == "" {
		name = "document"
	}

	replacer := strings.NewReplacer("/", "-", "\\", "-", "\"", "", "\n", " ", "\r", " ")
	name = strings.TrimSpace(replacer.Replace(name))
	if name == "" {
		name = "document"
	}

	return name + ".html"
}

func contentTypeOrDefault(value string) string {
	if strings.TrimSpace(value) == "" {
		return "text/html; charset=utf-8"
	}
	return value
}

var errEmptyMarkdown = errors.New("markdown must not be empty")

type errMarkdownTooLarge struct {
	limit int64
}

func (e errMarkdownTooLarge) Error() string {
	return fmt.Sprintf("markdown exceeds %d bytes", e.limit)
}
