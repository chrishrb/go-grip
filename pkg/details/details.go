// Package details provides a Goldmark extension for stateful collapsible details elements.
// It adds unique IDs to <details> elements and includes JavaScript to save/restore their
// state using browser session storage.
package details

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Extender implements goldmark.Extender to add stateful details support
type Extender struct {
	// IDPrefix is the prefix used for generated IDs. Defaults to "details-"
	IDPrefix string
}

// Extend extends the Goldmark parser and renderer with stateful details functionality
func (e *Extender) Extend(m goldmark.Markdown) {
	prefix := e.IDPrefix
	if prefix == "" {
		prefix = "details-"
	}

	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(NewTransformer(prefix), 100),
		),
	)

	// Enable unsafe HTML to allow <details> tags to be rendered
	m.Renderer().AddOptions(
		html.WithUnsafe(),
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(), 100),
		),
	)
}

// New creates a new details Extender with default settings
func New() *Extender {
	return &Extender{
		IDPrefix: "details-",
	}
}

// NewWithPrefix creates a new details Extender with a custom ID prefix
func NewWithPrefix(prefix string) *Extender {
	return &Extender{
		IDPrefix: prefix,
	}
}
