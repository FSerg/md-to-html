package converter

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const documentLang = "ru"

type Result struct {
	HTML  []byte
	Title string
}

type Converter struct {
	md         goldmark.Markdown
	tmpl       *template.Template
	bufferPool sync.Pool
}

type templateData struct {
	Lang      string
	Title     string
	Body      template.HTML
	ShowTitle bool
}

func New(templateFS fs.FS) (*Converter, error) {
	tmpl, err := template.ParseFS(templateFS, "document.html")
	if err != nil {
		return nil, err
	}

	return &Converter{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Footnote,
				emoji.Emoji,
				highlighting.NewHighlighting(
					highlighting.WithStyle("github"),
					highlighting.WithFormatOptions(chromahtml.WithClasses(false)),
				),
				&anchorExtension{},
			),
			goldmark.WithRendererOptions(
				renderer.WithNodeRenderers(
					util.Prioritized(&escapedRawHTMLRenderer{}, 999),
				),
			),
		),
		tmpl: tmpl,
		bufferPool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}, nil
}

func (c *Converter) Convert(md []byte, fallbackTitle string) (Result, error) {
	body, title, hasH1, err := c.render(md)
	if err != nil {
		return Result{}, err
	}

	if title == "" {
		title = fallbackTitle
	}

	buf := c.getBuffer()
	defer c.putBuffer(buf)

	data := templateData{
		Lang:      documentLang,
		Title:     title,
		Body:      template.HTML(body),
		ShowTitle: !hasH1 && title != "",
	}

	if err := c.tmpl.Execute(buf, data); err != nil {
		return Result{}, err
	}

	return Result{
		HTML:  append([]byte(nil), buf.Bytes()...),
		Title: title,
	}, nil
}

func (c *Converter) RenderBody(md []byte) ([]byte, string, error) {
	body, title, _, err := c.render(md)
	if err != nil {
		return nil, "", err
	}
	return body, title, nil
}

func (c *Converter) render(md []byte) ([]byte, string, bool, error) {
	root := c.md.Parser().Parse(text.NewReader(md))
	doc, ok := root.(*ast.Document)
	if !ok {
		return nil, "", false, fmt.Errorf("expected *ast.Document, got %T", root)
	}
	title, hasH1 := extractDocumentTitle(doc, md)

	buf := c.getBuffer()
	defer c.putBuffer(buf)

	if err := c.md.Renderer().Render(buf, md, doc); err != nil {
		return nil, "", false, err
	}

	return append([]byte(nil), buf.Bytes()...), title, hasH1, nil
}

func extractDocumentTitle(doc *ast.Document, src []byte) (string, bool) {
	var (
		title string
		hasH1 bool
	)

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		if h.Level == 1 {
			hasH1 = true
		}

		if title == "" {
			title = strings.TrimSpace(extractHeadingText(h, src))
		}

		return ast.WalkContinue, nil
	})

	return title, hasH1
}

func (c *Converter) getBuffer() *bytes.Buffer {
	buf := c.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (c *Converter) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	c.bufferPool.Put(buf)
}
