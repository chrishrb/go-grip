package pkg

import (
	"io"
	"log"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func (client *Client) MdToHTML(bytes []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(bytes)

	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags, RenderNodeHook: client.renderHookCodeBlock}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func (client *Client) renderHookCodeBlock(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	block, ok := node.(*ast.CodeBlock)
	if !ok {
		return ast.GoToNext, false
	}

	var style string
	switch client.Dark {
	case true:
		style = "github-dark"
	default:
		style = "github"
	}

	err := quick.Highlight(w, string(block.Literal), string(block.Info), "html", style)
	if err != nil {
		log.Println("Error:", err)
	}

	return ast.GoToNext, true
}
