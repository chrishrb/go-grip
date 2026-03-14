package mermaid

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func TestTransformer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc       string
		give       string
		noScript   bool
		wantBodies []string
		wantScript bool
	}{
		{
			desc: "empty",
			give: "",
		},
		{
			desc: "mermaid",
			give: unlines(
				"```mermaid",
				"foo",
				"```",
			),
			wantBodies: []string{"foo\n"},
			wantScript: true,
		},
		{
			desc: "mermaid and not",
			give: unlines(
				"Foo",
				"",
				"```mermaid",
				"foo",
				"```",
				"",
				"Bar",
				"",
				"```go",
				"bar",
				"",
				"Baz",
				"",
			),
			wantBodies: []string{"foo\n"},
			wantScript: true,
		},
		{
			desc:     "noscript",
			noScript: true,
			give: unlines(
				"```mermaid",
				"foo",
				"```",
			),
			wantBodies: []string{"foo\n"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			p := goldmark.New().Parser()
			p.AddOptions(
				parser.WithASTTransformers(
					util.Prioritized(&Transformer{
						NoScript: tt.noScript,
					}, 100),
				),
			)

			src := []byte(tt.give)
			got := p.Parse(text.NewReader(src))

			var (
				gotBodies []string
				gotScript int
			)
			err := ast.Walk(got, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
				if !enter {
					return ast.WalkContinue, nil
				}

				switch n := node.(type) {
				case *Block:
					var buff bytes.Buffer
					lines := n.Lines()
					for i := 0; i < lines.Len(); i++ {
						line := lines.At(i)
						buff.Write(line.Value(src))
					}

					gotBodies = append(gotBodies, buff.String())

				case *ScriptBlock:
					gotScript++
				}

				return ast.WalkContinue, nil
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantBodies, gotBodies)
			if tt.wantScript {
				assert.Equal(t, 1, gotScript)
			} else {
				assert.Zero(t, gotScript)
			}
		})
	}
}

func TestTransformer_RepeatedTransformations(t *testing.T) {
	t.Parallel()

	src := []byte(unlines(
		"```mermaid",
		"foo",
		"```",
	))
	r := text.NewReader(src)

	pctx := parser.NewContext()
	doc := goldmark.New().Parser().
		Parse(r, parser.WithContext(pctx)).(*ast.Document)

	var trans Transformer
	for i := 0; i < 10; i++ {
		trans.Transform(doc, r, pctx)
	}

	var scriptCount int
	err := ast.Walk(doc, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
		if _, ok := node.(*ScriptBlock); ok && enter {
			scriptCount++
		}
		return ast.WalkContinue, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, scriptCount)
}
