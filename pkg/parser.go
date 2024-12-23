package pkg

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"path"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/chrishrb/go-grip/defaults"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var blockquotes = []string{"Note", "Tip", "Important", "Warning", "Caution"}

func (client *Client) MdToHTML(bytes []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(bytes)

	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags, RenderNodeHook: client.renderHook}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func (client *Client) renderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch node.(type) {
	case *ast.BlockQuote:
		return renderHookBlockquote(w, node, entering)
	case *ast.Text:
		return renderHookText(w, node)
	case *ast.ListItem:
		return renderHookListItem(w, node, entering)
	case *ast.CodeBlock:
		return renderHookCodeBlock(w, node, client.Dark)
	}

	return ast.GoToNext, false
}

func renderHookCodeBlock(w io.Writer, node ast.Node, dark bool) (ast.WalkStatus, bool) {
	block := node.(*ast.CodeBlock)

	var style string
	switch dark {
	case true:
		style = "github-dark"
	default:
		style = "github"
	}

	if string(block.Info) == "mermaid" {
		m, err := renderMermaid(string(block.Literal), dark)
		if err != nil {
			log.Println("Error:", err)
		}
		fmt.Fprint(w, m)
		return ast.GoToNext, true
	}

	err := quick.Highlight(w, string(block.Literal), string(block.Info), "html", style)
	if err != nil {
		log.Println("Error:", err)
	}
	return ast.GoToNext, true
}

func renderHookBlockquote(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	block := node.(*ast.BlockQuote)

	paragraph, ok := (block.GetChildren()[0]).(*ast.Paragraph)
	if !ok {
		return ast.GoToNext, false
	}

	t, ok := (paragraph.GetChildren()[0]).(*ast.Text)
	if !ok {
		return ast.GoToNext, false
	}

	// Get the text content of the blockquote
	content := string(t.Literal)

	var alert string
	for _, b := range blockquotes {
		if strings.HasPrefix(content, fmt.Sprintf("[!%s]", strings.ToUpper(b))) {
			alert = strings.ToLower(b)
		}
	}

	if alert == "" {
		return ast.GoToNext, false
	}

	// Set the message type based on the content of the blockquote
	var err error
	if entering {
		s, _ := createBlockquoteStart(alert)
		_, err = io.WriteString(w, s)
	} else {
		_, err = io.WriteString(w, "</div>")
	}
	if err != nil {
		log.Println("Error:", err)
	}

	return ast.GoToNext, true
}

func renderHookText(w io.Writer, node ast.Node) (ast.WalkStatus, bool) {
	block := node.(*ast.Text)

	r := regexp.MustCompile(`(:\S+:)`)
	withEmoji := r.ReplaceAllStringFunc(string(block.Literal), func(s string) string {
		val, ok := EmojiMap[s]
		if !ok {
			return s
		}

		if strings.HasPrefix(val, "/") {
			return fmt.Sprintf(`<img class="emoji" title="%s" alt="%s" src="%s" height="20" width="20" align="absmiddle">`, s, s, val)
		}

		return val
	})

	paragraph, ok := block.GetParent().(*ast.Paragraph)
	if !ok {
		_, err := io.WriteString(w, withEmoji)
		if err != nil {
			log.Println("Error:", err)
		}
		return ast.GoToNext, true
	}

	_, ok = paragraph.GetParent().(*ast.BlockQuote)
	if ok {
		// Remove prefixes
		for _, b := range blockquotes {
			content, found := strings.CutPrefix(withEmoji, fmt.Sprintf("[!%s]", strings.ToUpper(b)))
			if found {
				_, err := io.WriteString(w, content)
				if err != nil {
					log.Println("Error:", err)
				}
				return ast.GoToNext, true
			}
		}
	}

	_, ok = paragraph.GetParent().(*ast.ListItem)
	if ok {
		content, found := strings.CutPrefix(withEmoji, "[ ]")
		content = `<input type="checkbox" disabled class="task-list-item-checkbox"> ` + content
		if found {
			_, err := io.WriteString(w, content)
			if err != nil {
				log.Println("Error:", err)
			}
			return ast.GoToNext, true
		}

		content, found = strings.CutPrefix(withEmoji, "[x]")
		content = `<input type="checkbox" disabled class="task-list-item-checkbox" checked> ` + content
		if found {
			_, err := io.WriteString(w, content)
			if err != nil {
				log.Println("Error:", err)
			}
		}
	}

	_, err := io.WriteString(w, withEmoji)
	if err != nil {
		log.Println("Error:", err)
	}
	return ast.GoToNext, true
}

func renderHookListItem(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	block := node.(*ast.ListItem)

	paragraph, ok := (block.GetChildren()[0]).(*ast.Paragraph)
	if !ok {
		return ast.GoToNext, false
	}

	t, ok := (paragraph.GetChildren()[0]).(*ast.Text)
	if !ok {
		return ast.GoToNext, false
	}

	if !(strings.HasPrefix(string(t.Literal), "[ ]") || strings.HasPrefix(string(t.Literal), "[x]")) {
		return ast.GoToNext, false
	}

	if entering {
		_, err := io.WriteString(w, "<li class=\"task-list-item\">")
		if err != nil {
			log.Println("Error:", err)
		}
	} else {
		_, err := io.WriteString(w, "</li>")
		if err != nil {
			log.Println("Error:", err)
		}
	}

	return ast.GoToNext, true
}

func createBlockquoteStart(alert string) (string, error) {
	lp := path.Join("templates/alert", fmt.Sprintf("%s.html", alert))
	tmpl, err := template.ParseFS(defaults.Templates, lp)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, alert); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

type mermaid struct {
	Content  string
	Darkmode bool
}

func renderMermaid(content string, darkmode bool) (string, error) {
	m := mermaid{
		Content:  content,
		Darkmode: darkmode,
	}
	lp := path.Join("templates/mermaid/mermaid.html")
	tmpl, err := template.ParseFS(defaults.Templates, lp)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, m); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
