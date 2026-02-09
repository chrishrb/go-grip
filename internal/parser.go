package internal

import (
	"bytes"

	"github.com/chrishrb/go-grip/pkg/alert"
	"github.com/chrishrb/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/hashtag"
)

type Parser struct {
	theme string
}

func NewParser(theme string) *Parser {
	return &Parser{
		theme: theme,
	}
}

func (m Parser) MdToHTML(input []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Table,
			extension.Strikethrough,
			tasklist.TaskList,
			emoji.Emoji,
			&hashtag.Extender{},
			alert.New(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(input, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
