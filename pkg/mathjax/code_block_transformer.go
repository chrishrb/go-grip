package mathjax

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type mathCodeBlockTransformer struct {
}

var defaultMathCodeBlockTransformer = &mathCodeBlockTransformer{}

func NewMathCodeBlockTransformer() parser.ASTTransformer {
	return defaultMathCodeBlockTransformer
}

func (t *mathCodeBlockTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Check if this is a fenced code block
		if codeBlock, ok := n.(*ast.FencedCodeBlock); ok {
			// Check if the language is "math"
			if codeBlock.Info != nil {
				language := codeBlock.Info.Text(reader.Source())
				if bytes.Equal(language, []byte("math")) {
					// Convert to MathBlock
					mathBlock := NewMathBlock()
					
					// Copy all lines from the code block to the math block
					for i := 0; i < codeBlock.Lines().Len(); i++ {
						line := codeBlock.Lines().At(i)
						mathBlock.Lines().Append(line)
					}
					
					// Replace the code block with the math block
					parent := n.Parent()
					if parent != nil {
						parent.ReplaceChild(parent, codeBlock, mathBlock)
					}
				}
			}
		}

		return ast.WalkContinue, nil
	})
}
