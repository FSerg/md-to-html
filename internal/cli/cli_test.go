package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFileToFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "example.md")
	if err := os.WriteFile(inputPath, []byte("# Hello\n\nBody"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if err := Run(context.Background(), []string{inputPath}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("run: %v", err)
	}

	outputPath := filepath.Join(dir, "example.html")
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Contains(got, []byte("<!DOCTYPE html>")) {
		t.Fatalf("output missing doctype: %s", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestCLIStdin(t *testing.T) {
	t.Parallel()

	stdin := strings.NewReader("# Привет\n\nТекст")
	var stdout, stderr bytes.Buffer

	if err := Run(context.Background(), []string{"--stdin"}, stdin, &stdout, &stderr); err != nil {
		t.Fatalf("run: %v", err)
	}

	if !strings.Contains(stdout.String(), "<!DOCTYPE html>") {
		t.Fatalf("stdout missing doctype: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestCLIOutputFlag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "example.md")
	outputPath := filepath.Join(dir, "custom.html")
	if err := os.WriteFile(inputPath, []byte("Plain text"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if err := Run(context.Background(), []string{inputPath, "-o", outputPath}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("stat output: %v", err)
	}
}

func TestCLITitle(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if err := Run(context.Background(), []string{"--stdin", "--title", "Custom"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("run: %v", err)
	}

	if !strings.Contains(stdout.String(), "<title>Custom</title>") {
		t.Fatalf("stdout missing title: %s", stdout.String())
	}
}

func TestCLINoInput(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), nil, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("error = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "Usage: md-to-html cli") {
		t.Fatalf("stderr missing usage: %s", stderr.String())
	}
}

func TestCLIMissingFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"missing.md"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrUsage) {
		t.Fatalf("error = %v, did not want ErrUsage", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}
