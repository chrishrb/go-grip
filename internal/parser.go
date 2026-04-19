package internal

import (
	"bytes"

	"github.com/chrishrb/go-grip/pkg/alert"
	"github.com/chrishrb/go-grip/pkg/details"
	"github.com/chrishrb/go-grip/pkg/footnote"
	"github.com/chrishrb/go-grip/pkg/ghissue"
	"github.com/chrishrb/go-grip/pkg/highlighting"
	"github.com/chrishrb/go-grip/pkg/mathjax"
	"github.com/chrishrb/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (m Parser) MdToHTML(input []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Table,
			extension.Strikethrough,
			footnote.Footnote,
			tasklist.TaskList,
			emoji.Emoji,
			&hashtag.Extender{},
			alert.New(),
			highlighting.Highlighting,
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
			mathjax.MathJax,
			ghissue.New(),
			details.New(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(input, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
