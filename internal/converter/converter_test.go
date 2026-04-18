package converter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fserg/md-to-html/web/template"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestGolden(t *testing.T) {
	c := newTestConverter(t)
	update := os.Getenv("UPDATE_GOLDEN") == "1"

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			md, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}

			wantPath := filepath.Join("testdata", strings.TrimSuffix(name, ".md")+".html")
			got, err := c.Convert(md, "Document")
			if err != nil {
				t.Fatal(err)
			}

			for _, forbidden := range []string{"http://", "https://", "cdn.", "googleapis.com"} {
				if bytes.Contains(got.HTML, []byte(forbidden)) {
					t.Fatalf("generated HTML contains forbidden external resource marker %q", forbidden)
				}
			}

			if update {
				if err := os.WriteFile(wantPath, got.HTML, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("missing golden %s; run UPDATE_GOLDEN=1", wantPath)
			}

			if !bytes.Equal(got.HTML, want) {
				t.Errorf("mismatch: run UPDATE_GOLDEN=1 go test ./internal/converter/... to refresh")
			}
		})
	}
}

func TestTranslitSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		used map[string]int
	}{
		{name: "cyrillic", in: "Установка", want: "ustanovka", used: map[string]int{}},
		{name: "collision first", in: "Install", want: "install", used: map[string]int{}},
		{name: "collision second", in: "Install", want: "install-1", used: map[string]int{"install": 1}},
		{name: "cyrillic translit", in: "Сетап", want: "setap", used: map[string]int{}},
		{name: "empty fallback", in: "!!!", want: "section", used: map[string]int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translitSlug(tt.in, tt.used)
			if got != tt.want {
				t.Fatalf("translitSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractHeadingText(t *testing.T) {
	c := newTestConverter(t)
	src := []byte("## [API](https://example.com) `go fmt` https://example.com :rocket:\n")
	doc := c.md.Parser().Parse(text.NewReader(src))

	var heading *ast.Heading
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			heading = h
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})

	if heading == nil {
		t.Fatal("heading not found")
	}

	got := extractHeadingText(heading, src)
	want := "API go fmt https://example.com 🚀"
	if got != want {
		t.Fatalf("extractHeadingText() = %q, want %q", got, want)
	}
}

func TestConvertTitleFromFirstHeading(t *testing.T) {
	c := newTestConverter(t)

	result, err := c.Convert([]byte("# Hello\n\nParagraph"), "fallback")
	if err != nil {
		t.Fatal(err)
	}

	if result.Title != "Hello" {
		t.Fatalf("result.Title = %q, want %q", result.Title, "Hello")
	}

	if !bytes.Contains(result.HTML, []byte("<title>Hello</title>")) {
		t.Fatalf("expected HTML title to contain Hello")
	}
}

func TestConvertTitleFallback(t *testing.T) {
	c := newTestConverter(t)

	result, err := c.Convert([]byte("Paragraph only"), "fallback")
	if err != nil {
		t.Fatal(err)
	}

	if result.Title != "fallback" {
		t.Fatalf("result.Title = %q, want %q", result.Title, "fallback")
	}

	if !bytes.Contains(result.HTML, []byte("<h1>fallback</h1>")) {
		t.Fatalf("expected fallback h1 to be injected")
	}
}

func newTestConverter(t *testing.T) *Converter {
	t.Helper()

	c, err := New(webtemplate.FS)
	if err != nil {
		t.Fatal(err)
	}

	return c
}
