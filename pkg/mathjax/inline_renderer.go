package mathjax

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type InlineMathRenderer struct {
	startDelim string
	endDelim   string
}

func (r *InlineMathRenderer) renderInlineMath(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<span class="math inline">` + r.startDelim)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				_, _ = w.Write(value[:len(value)-1])
				if c != n.LastChild() {
					_, _ = w.Write([]byte(" "))
				}
			} else {
				_, _ = w.Write(value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(r.endDelim + `</span>`)
	return ast.WalkContinue, nil
}

func (r *InlineMathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInlineMath, r.renderInlineMath)
}
