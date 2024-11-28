package handlers

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

func RenderMarkdown(content []byte) (template.HTML, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	context := parser.NewContext()
	var buf bytes.Buffer

	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
