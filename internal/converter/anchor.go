package converter

import (
	"strings"

	"github.com/yuin/goldmark"
	emojiast "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type anchorExtension struct{}

func (e *anchorExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&anchorTransformer{}, 900),
	))
}

type anchorTransformer struct{}

func (t *anchorTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	src := reader.Source()
	used := map[string]int{}

	_ = pc

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		slug := translitSlug(extractHeadingText(h, src), used)
		h.SetAttributeString("id", []byte(slug))

		link := ast.NewLink()
		link.Destination = []byte("#" + slug)
		link.SetAttributeString("class", []byte("heading-anchor"))
		link.SetAttributeString("aria-hidden", []byte("true"))
		link.AppendChild(link, ast.NewString([]byte("#")))

		if first := h.FirstChild(); first != nil {
			h.InsertBefore(h, first, link)
		} else {
			h.AppendChild(h, link)
		}

		return ast.WalkSkipChildren, nil
	})
}

func extractHeadingText(h *ast.Heading, src []byte) string {
	var b strings.Builder

	_ = ast.Walk(h, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Link:
			if isHeadingAnchor(v) {
				return ast.WalkSkipChildren, nil
			}
		case *ast.Text:
			b.Write(v.Segment.Value(src))
			if v.HardLineBreak() || v.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(v.Value)
		case *ast.CodeSpan:
			for child := v.FirstChild(); child != nil; child = child.NextSibling() {
				switch c := child.(type) {
				case *ast.Text:
					b.Write(c.Segment.Value(src))
				case *ast.String:
					b.Write(c.Value)
				}
			}
			return ast.WalkSkipChildren, nil
		case *ast.AutoLink:
			b.Write(v.Label(src))
			return ast.WalkSkipChildren, nil
		case *emojiast.Emoji:
			if v.Value != nil && len(v.Value.Unicode) > 0 {
				b.WriteString(string(v.Value.Unicode))
			} else if len(v.ShortName) > 0 {
				b.WriteByte(':')
				b.Write(v.ShortName)
				b.WriteByte(':')
			}
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	return strings.TrimSpace(b.String())
}

func isHeadingAnchor(link *ast.Link) bool {
	attr, ok := link.AttributeString("class")
	if !ok {
		return false
	}

	switch value := attr.(type) {
	case []byte:
		return string(value) == "heading-anchor"
	case string:
		return value == "heading-anchor"
	default:
		return false
	}
}
