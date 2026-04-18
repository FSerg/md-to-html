//go:build tools

package tools

import (
	_ "github.com/a-h/templ"
	_ "github.com/alecthomas/chroma/v2"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/google/uuid"
	_ "github.com/mozillazg/go-unidecode"
	_ "github.com/yuin/goldmark"
	_ "github.com/yuin/goldmark-emoji"
	_ "github.com/yuin/goldmark-highlighting/v2"
)
