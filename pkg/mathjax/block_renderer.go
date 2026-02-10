package mathjax

import (
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type MathBlockRenderer struct {
	startDelim string
	endDelim   string
}

func NewMathBlockRenderer(start, end string) renderer.NodeRenderer {
	return &MathBlockRenderer{start, end}
}

func (r *MathBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathBlock, r.renderMathBlock)
}

func (r *MathBlockRenderer) writeLines(w util.BufWriter, source []byte, n gast.Node) bool {
	l := n.Lines().Len()
	endsWithNewline := false
	for i := range l {
		line := n.Lines().At(i)
		lineBytes := line.Value(source)
		w.Write(lineBytes)
		if len(lineBytes) > 0 && lineBytes[len(lineBytes)-1] == '\n' {
			endsWithNewline = true
		} else {
			endsWithNewline = false
		}
	}
	return endsWithNewline
}

func (r *MathBlockRenderer) renderMathBlock(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*MathBlock)
	if entering {
		_, _ = w.WriteString(`<p><span class="math display">` + r.startDelim)
		endsWithNewline := r.writeLines(w, source, n)
		// Add a newline before the closing delimiter if content doesn't end with one
		if !endsWithNewline {
			_, _ = w.WriteString("\n")
		}
	} else {
		_, _ = w.WriteString(r.endDelim + `</span></p>` + "\n")
	}
	return gast.WalkContinue, nil
}
