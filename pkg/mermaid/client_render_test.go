package mermaid

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func TestRenderer_Block(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give string

		tag string // ContainerTag option

		want string
	}{
		{
			desc: "empty",
			give: "",
			want: `<pre class="mermaid"></pre>`,
		},
		{
			desc: "graph",
			give: "graph TD;",
			want: `<pre class="mermaid">graph TD;</pre>`,
		},
		{
			desc: "newlines",
			give: unlines("foo", "bar"),
			want: `<pre class="mermaid">foo` + "\nbar" + "\n</pre>",
		},
		{
			desc: "escaping",
			give: "A -> B",
			want: `<pre class="mermaid">A -&gt; B</pre>`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			r := buildNodeRenderer(&ClientRenderer{})

			reader := text.NewReader([]byte(tt.give))
			give := blockFromReader(reader)

			var buff bytes.Buffer
			assert.NoError(t, r.Render(&buff, reader.Source(), give), "Render")
			assert.Equal(t, tt.want, buff.String())
		})
	}
}

func TestRenderer_Script(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc      string
		mermaidJS string
		want      string
	}{
		{
			desc: "default mermaid.js",
			want: fmt.Sprintf(`<script src=%q></script><script>mermaid.initialize({"startOnLoad":true});</script>`, _defaultMermaidJS),
		},
		{
			desc:      "explicit mermaid.js",
			mermaidJS: "mermaid.js",
			want:      `<script src="mermaid.js"></script><script>mermaid.initialize({"startOnLoad":true});</script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			r := buildNodeRenderer(&ClientRenderer{
				MermaidURL: tt.mermaidJS,
			})

			var buff bytes.Buffer
			assert.NoError(t,
				r.Render(&buff, nil /* src */, &ScriptBlock{}))
			assert.Equal(t, tt.want, buff.String())
		})
	}
}

func TestRenderer_Script_withOptions(t *testing.T) {
	t.Parallel()

	r := buildNodeRenderer(&ClientRenderer{
		MermaidURL: "mermaid.js",
	})

	var buff bytes.Buffer
	assert.NoError(t,
		r.Render(&buff, nil /* src */, &ScriptBlock{}))
	assert.Equal(t,
		`<script src="mermaid.js"></script><script>mermaid.initialize({"startOnLoad":true,"theme":"dark"});</script>`,
		buff.String())
}

func buildNodeRenderer(r renderer.NodeRenderer) renderer.Renderer {
	return renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(r, 100),
		),
	)
}
