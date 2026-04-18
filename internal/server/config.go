package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr             = ":8080"
	defaultMaxMarkdownBytes = int64(1_048_576)
	defaultMaxRequestBytes  = int64(1_200_000)
	defaultPreviewTTL       = time.Hour
	defaultShutdownTimeout  = 10 * time.Second
)

type Config struct {
	Addr             string
	MaxMarkdownBytes int64
	MaxRequestBytes  int64
	PreviewTTL       time.Duration
	ShutdownTimeout  time.Duration
}

func LoadConfig() (Config, error) {
	maxMarkdownBytes, err := loadPositiveInt64("MAX_MARKDOWN_BYTES", defaultMaxMarkdownBytes)
	if err != nil {
		return Config{}, err
	}

	maxRequestBytes, err := loadPositiveInt64("MAX_REQUEST_BYTES", defaultMaxRequestBytes)
	if err != nil {
		return Config{}, err
	}

	previewTTL, err := loadDuration("PREVIEW_TTL", defaultPreviewTTL)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := loadDuration("SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	addr := strings.TrimSpace(os.Getenv("ADDR"))
	if addr == "" {
		addr = defaultAddr
	}

	return Config{
		Addr:             addr,
		MaxMarkdownBytes: maxMarkdownBytes,
		MaxRequestBytes:  maxRequestBytes,
		PreviewTTL:       previewTTL,
		ShutdownTimeout:  shutdownTimeout,
	}, nil
}

func loadPositiveInt64(name string, fallback int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return value, nil
}

func loadDuration(name string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return value, nil
}
