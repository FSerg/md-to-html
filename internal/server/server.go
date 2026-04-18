package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fserg/md-to-html/internal/converter"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func New(cfg Config, conv *converter.Converter) (*Server, error) {
	if conv == nil {
		return nil, errors.New("converter is required")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &Server{
		cfg:   cfg,
		conv:  conv,
		store: NewPreviewStore(cfg.PreviewTTL),
		log:   logger,
	}, nil
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(CORSMiddleware())
	r.Use(MaxBytesMiddleware(s.cfg.MaxRequestBytes))
	r.Use(RequestLogger(s.log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/", s.handleHome)
	r.Post("/convert", s.handleConvert)
	r.Get("/health", s.handleHealth)
	r.Get("/version", s.handleVersion)
	r.Get("/ready", s.handleReady)
	r.Post("/ui/convert", s.handleUIConvert)
	r.Get("/preview/{id}", s.handlePreview)
	r.Get("/download/{id}", s.handleDownload)

	return r
}

func (s *Server) Run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    s.cfg.Addr,
		Handler: s.Router(),
	}

	errCh := make(chan error, 1)

	go s.store.janitor(ctx)
	go func() {
		s.log.Info("server starting", "addr", s.cfg.Addr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down", "timeout", s.cfg.ShutdownTimeout)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server exited after shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	}
}
