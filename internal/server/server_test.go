package server

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/fserg/md-to-html/internal/converter"
	"github.com/fserg/md-to-html/internal/version"
	webtemplate "github.com/fserg/md-to-html/web/template"
)

func TestConvertEndpoint(t *testing.T) {
	srv := newTestServer(t, Config{
		Addr:             ":0",
		MaxMarkdownBytes: 128,
		MaxRequestBytes:  256,
		PreviewTTL:       time.Hour,
		ShutdownTimeout:  time.Second,
	})

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	tests := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
		wantType    string
		wantBody    string
	}{
		{
			name:        "valid markdown",
			body:        `{"markdown":"# Hello"}`,
			contentType: "application/json",
			wantStatus:  http.StatusOK,
			wantType:    "text/html; charset=utf-8",
			wantBody:    "<!DOCTYPE html>",
		},
		{
			name:        "empty markdown",
			body:        `{"markdown":"   "}`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
			wantType:    "application/json; charset=utf-8",
			wantBody:    `{"detail":"markdown must not be empty"}`,
		},
		{
			name:        "markdown too large",
			body:        `{"markdown":"` + strings.Repeat("a", 129) + `"}`,
			contentType: "application/json",
			wantStatus:  http.StatusRequestEntityTooLarge,
			wantType:    "application/json; charset=utf-8",
			wantBody:    `{"detail":"markdown exceeds 128 bytes"}`,
		},
		{
			name:        "missing content type",
			body:        `{"markdown":"# Hello"}`,
			contentType: "",
			wantStatus:  http.StatusUnsupportedMediaType,
			wantType:    "application/json; charset=utf-8",
			wantBody:    `{"detail":"content-type must be application/json"}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/convert", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, tc.wantStatus, body)
			}
			if got := resp.Header.Get("Content-Type"); got != tc.wantType {
				t.Fatalf("content-type = %q, want %q", got, tc.wantType)
			}
			if !bytes.Contains(body, []byte(tc.wantBody)) {
				t.Fatalf("body %q does not contain %q", body, tc.wantBody)
			}
		})
	}
}

func TestConvertEndpoint_RequestLimit(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, Config{
		Addr:             ":0",
		MaxMarkdownBytes: 1_048_576,
		MaxRequestBytes:  64,
		PreviewTTL:       time.Hour,
		ShutdownTimeout:  time.Second,
	})

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/convert", strings.NewReader(`{"markdown":"`+strings.Repeat("a", 100)+`"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, http.StatusRequestEntityTooLarge, body)
	}
	if !bytes.Contains(body, []byte(`{"detail":"request exceeds 64 bytes"}`)) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestStatusEndpoints(t *testing.T) {
	originalVersion := version.Version
	version.Version = "dev"
	t.Cleanup(func() {
		version.Version = originalVersion
	})

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	tests := []struct {
		path string
		want map[string]any
	}{
		{path: "/health", want: map[string]any{"status": "ok"}},
		{path: "/version", want: map[string]any{"version": "dev"}},
		{path: "/ready", want: map[string]any{"status": "ok", "template_loaded": true}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			resp, err := ts.Client().Get(ts.URL + tc.path)
			if err != nil {
				t.Fatalf("get %s: %v", tc.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			var got map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			for key, wantValue := range tc.want {
				if got[key] != wantValue {
					t.Fatalf("%s[%q] = %v, want %v", tc.path, key, got[key], wantValue)
				}
			}
		})
	}
}

func TestHomePage(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get home: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read home body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want %q", got, "text/html; charset=utf-8")
	}
	for _, needle := range []string{
		`hx-post="/ui/convert"`,
		`id="result"`,
		`value="file"`,
		`value="text"`,
	} {
		if !bytes.Contains(body, []byte(needle)) {
			t.Fatalf("home body missing %q", needle)
		}
	}
}

func TestUIConvertWithText(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	body, contentType := newMultipartRequest(t, map[string]string{
		"source":        "text",
		"markdown_text": "# Привет мир\n\nТекст",
	}, nil)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/ui/convert", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, http.StatusOK, respBody)
	}
	for _, needle := range []string{
		"Открыть превью",
		"Скачать HTML",
		`/preview/`,
		`/download/`,
		`srcdoc=`,
		`document.html`,
	} {
		if !bytes.Contains(respBody, []byte(needle)) {
			t.Fatalf("response missing %q", needle)
		}
	}
}

func TestUIConvertWithFile(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	body, contentType := newMultipartRequest(t, map[string]string{
		"source": "file",
	}, map[string]filePart{
		"markdown_file": {
			filename: "guide.md",
			content:  "# Guide\n\nBody",
		},
	})

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/ui/convert", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, http.StatusOK, respBody)
	}
	if !bytes.Contains(respBody, []byte("guide.html")) {
		t.Fatalf("response missing filename; body=%s", respBody)
	}
}

func TestUIConvertErrors(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, Config{
		Addr:             ":0",
		MaxMarkdownBytes: 8,
		MaxRequestBytes:  1024,
		PreviewTTL:       time.Hour,
		ShutdownTimeout:  time.Second,
	})
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	tests := []struct {
		name       string
		fields     map[string]string
		files      map[string]filePart
		wantStatus int
		wantBody   string
	}{
		{
			name:       "empty text",
			fields:     map[string]string{"source": "text", "markdown_text": "   "},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Пустой markdown",
		},
		{
			name:       "missing file",
			fields:     map[string]string{"source": "file"},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Файл не загружен",
		},
		{
			name:       "markdown too large",
			fields:     map[string]string{"source": "text", "markdown_text": strings.Repeat("x", 9)},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   "Markdown больше 8 байт",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			body, contentType := newMultipartRequest(t, tc.fields, tc.files)

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/ui/convert", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.Header.Set("Content-Type", contentType)

			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, tc.wantStatus, respBody)
			}
			if !bytes.Contains(respBody, []byte(tc.wantBody)) {
				t.Fatalf("response %q missing %q", respBody, tc.wantBody)
			}
		})
	}
}

func TestPreviewAndDownloadOneShot(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	previewID := srv.store.Put([]byte("<h1>Preview</h1>"), "text/html; charset=utf-8", "preview.html")
	downloadID := srv.store.Put([]byte("<h1>Download</h1>"), "text/html; charset=utf-8", "download.html")

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/preview/" + previewID)
	if err != nil {
		t.Fatalf("get preview: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read preview body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("preview status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("preview cache-control = %q, want %q", got, "no-store")
	}
	if string(body) != "<h1>Preview</h1>" {
		t.Fatalf("preview body = %q", body)
	}

	resp, err = ts.Client().Get(ts.URL + "/preview/" + previewID)
	if err != nil {
		t.Fatalf("get preview second time: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second preview status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	resp, err = ts.Client().Get(ts.URL + "/download/" + downloadID)
	if err != nil {
		t.Fatalf("get download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, `attachment; filename=preview.html`) && !strings.Contains(got, `attachment; filename=download.html`) {
		t.Fatalf("unexpected content-disposition: %q", got)
	}

	resp, err = ts.Client().Get(ts.URL + "/download/" + downloadID)
	if err != nil {
		t.Fatalf("get download second time: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second download status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestPreviewMissing(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/preview/nonexistent")
	if err != nil {
		t.Fatalf("get preview: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestCORSPreflight(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, defaultTestConfig())
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/convert", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "https://evil.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin = %q, want %q", got, "*")
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "POST, GET, OPTIONS" {
		t.Fatalf("allow-methods = %q", got)
	}
}

func newTestServer(t *testing.T, cfg Config) *Server {
	t.Helper()

	conv, err := converter.New(webtemplate.FS)
	if err != nil {
		t.Fatalf("new converter: %v", err)
	}

	srv, err := New(cfg, conv)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	return srv
}

func defaultTestConfig() Config {
	return Config{
		Addr:             ":0",
		MaxMarkdownBytes: 1_048_576,
		MaxRequestBytes:  1_200_000,
		PreviewTTL:       time.Hour,
		ShutdownTimeout:  time.Second,
	}
}

type filePart struct {
	filename string
	content  string
}

func newMultipartRequest(t *testing.T, fields map[string]string, files map[string]filePart) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("write field %s: %v", name, err)
		}
	}

	for name, file := range files {
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", `form-data; name="`+name+`"; filename="`+file.filename+`"`)
		header.Set("Content-Type", "text/markdown")

		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("create part %s: %v", name, err)
		}
		if _, err := io.WriteString(part, file.content); err != nil {
			t.Fatalf("write part %s: %v", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return buf.Bytes(), writer.FormDataContentType()
}
