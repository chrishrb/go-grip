// Package alert provides a Goldmark extension for GitHub-style alerts.
// It transforms blockquotes with [!NOTE], [!TIP], [!IMPORTANT], [!WARNING], or [!CAUTION]
// into styled alert blocks matching GitHub's markdown rendering.
package alert

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Extender implements goldmark.Extender to add alert support
type Extender struct{}

// Extend extends the Goldmark parser and renderer with alert functionality
func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(NewTransformer(), 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(), 100),
		),
	)
}

// New creates a new alert Extender
func New() *Extender {
	return &Extender{}
}
