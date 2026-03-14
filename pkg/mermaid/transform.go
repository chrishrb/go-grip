package mermaid

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Transformer transforms a Goldmark Markdown AST with support for Mermaid
// diagrams. It makes the following transformations:
//
//   - replace mermaid code blocks with mermaid.Block nodes
//   - add a mermaid.ScriptBlock node if the document uses Mermaid
//     and one does not already exist
type Transformer struct {
	// Don't add a ScriptBlock to the end of the page
	// even if the page doesn't already have one.
	NoScript bool
}

var _mermaid = []byte("mermaid")

// Transform transforms the provided Markdown AST.
func (t *Transformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	var (
		hasScript     bool
		mermaidBlocks []*ast.FencedCodeBlock
	)

	// Collect all blocks to be replaced without modifying the tree.
	_ = ast.Walk(doc, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			return ast.WalkContinue, nil
		}

		// For multiple transforms.
		if _, ok := node.(*ScriptBlock); ok {
			hasScript = true
			return ast.WalkContinue, nil
		}

		cb, ok := node.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		lang := cb.Language(reader.Source())
		if !bytes.Equal(lang, _mermaid) {
			return ast.WalkContinue, nil
		}

		mermaidBlocks = append(mermaidBlocks, cb)
		return ast.WalkContinue, nil
	})

	// Nothing to do.
	if len(mermaidBlocks) == 0 {
		return
	}

	for _, cb := range mermaidBlocks {
		b := new(Block)
		b.SetLines(cb.Lines())

		parent := cb.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, cb, b)
		}
	}

	if !hasScript && !t.NoScript {
		doc.AppendChild(doc, &ScriptBlock{})
	}
}
