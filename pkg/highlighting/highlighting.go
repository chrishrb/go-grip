// MIT License
// Part of this code is forked from https://github.com/yuin/goldmark-highlighting
// Copyright (c) 2019 Yusuke Inuzuka

package highlighting

import (
	"bytes"
	"html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	goldmark_html "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HTMLRenderer struct is a renderer.NodeRenderer implementation for the extension.
type HTMLRenderer struct {
	goldmark_html.Config
}

// NewHTMLRenderer builds a new HTMLRenderer and returns it.
func NewHTMLRenderer() renderer.NodeRenderer {
	return &HTMLRenderer{
		Config: goldmark_html.NewConfig(),
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func renderClipboardButton(w util.BufWriter, code string) {
	// Escape the code for use in HTML attribute
	escapedCode := html.EscapeString(code)

	_, _ = w.WriteString(`<div class="zeroclipboard-container">
    <clipboard-copy aria-label="Copy" class="ClipboardButton btn btn-invisible js-clipboard-copy m-2 p-0 d-flex flex-justify-center flex-items-center" data-copy-feedback="Copied!" data-tooltip-direction="w" value="`)
	_, _ = w.WriteString(escapedCode)
	_, _ = w.WriteString(`" tabindex="0" role="button">
      <svg aria-hidden="true" height="16" viewBox="0 0 16 16" version="1.1" width="16" data-view-component="true" class="octicon octicon-copy js-clipboard-copy-icon">
    <path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 0 1 0 1.5h-1.5a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-1.5a.75.75 0 0 1 1.5 0v1.5A1.75 1.75 0 0 1 9.25 16h-7.5A1.75 1.75 0 0 1 0 14.25Z"></path><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0 1 14.25 11h-7.5A1.75 1.75 0 0 1 5 9.25Zm1.75-.25a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-7.5a.25.25 0 0 0-.25-.25Z"></path>
</svg>
      <svg aria-hidden="true" height="16" viewBox="0 0 16 16" version="1.1" width="16" data-view-component="true" class="octicon octicon-check js-clipboard-check-icon color-fg-success d-none">
    <path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"></path>
</svg>
    </clipboard-copy>
  </div>`)
}

func (r *HTMLRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if !entering {
		return ast.WalkContinue, nil
	}

	_, _ = w.WriteString(`<div class="highlight notranslate position-relative overflow-auto"">`)

	language := n.Language(source)
	var lexer chroma.Lexer
	if language != nil {
		lexer = lexers.Get(string(language))
	}
	if lexer == nil {
		lexer = lexers.Get("plaintext")
	}

	var buffer bytes.Buffer
	l := n.Lines().Len()
	for i := range l {
		line := n.Lines().At(i)
		buffer.Write(line.Value(source))
	}

	lexer = chroma.Coalesce(lexer)
	iterator, _ := lexer.Tokenise(nil, buffer.String())
	style := styles.Fallback
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	_ = formatter.Format(w, style, iterator)

	renderClipboardButton(w, buffer.String())

	_, _ = w.WriteString("</div>")
	return ast.WalkContinue, nil
}

type highlighting struct{}

// Highlighting is a goldmark.Extender implementation.
var Highlighting = &highlighting{}

// NewHighlighting returns a new extension.
func NewHighlighting() goldmark.Extender {
	return &highlighting{}
}

// Extend implements goldmark.Extender.
func (e *highlighting) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHTMLRenderer(), 200),
	))
}
