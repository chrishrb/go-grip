package mermaid

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type mermaid struct {
	// URL of Mermaid Javascript to be included in the page
	// for client-side rendering.
	//
	// Ignored if NoScript is true or if we're rendering diagrams server-side.
	//
	// Defaults to the latest version available on cdn.jsdelivr.net.
	MermaidURL string

	// If true, don't add a <script> including Mermaid to the end of the
	// page even if rendering diagrams client-side.
	//
	// Use this if the page you're including goldmark-mermaid in
	// already has a MermaidJS script included elsewhere.
	NoScript bool

	// Theme for mermaid diagrams.
	//
	// Values include "dark", "light" and "auto".
	Theme string
}

func NewMermaid() *mermaid {
	return &mermaid{}
}

// Extend extends the provided Goldmark parser with support for Mermaid
// diagrams.
func (e *mermaid) Extend(md goldmark.Markdown) {
	var themeVariables themeVariables
	if e.Theme == "dark" {
		themeVariables = darkThemeVariables
	} else {
		themeVariables = lightThemeVariables
	}

	r := &ClientRenderer{
		MermaidURL:     e.MermaidURL,
		ThemeVariables: themeVariables,
	}

	md.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&Transformer{}, 100),
		),
	)

	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(r, 100),
		),
	)
}
