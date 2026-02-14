package details

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Transformer is a transformer that adds IDs to HTML <details> elements
type Transformer struct {
	idPrefix string
	counter  int
}

// Transform walks the AST and modifies HTML blocks containing <details> tags
func (t *Transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	var toProcess []struct {
		htmlBlock *ast.HTMLBlock
		content   string
		id        string
	}

	// Collect all HTML blocks that start with <details
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if htmlBlock, ok := n.(*ast.HTMLBlock); ok {
			var htmlContent strings.Builder
			for i := 0; i < htmlBlock.Lines().Len(); i++ {
				line := htmlBlock.Lines().At(i)
				htmlContent.Write(line.Value(source))
			}

			content := htmlContent.String()
			contentTrimmed := strings.TrimSpace(content)

			if strings.HasPrefix(contentTrimmed, "<details") {
				t.counter++
				id := t.generateID(content)
				toProcess = append(toProcess, struct {
					htmlBlock *ast.HTMLBlock
					content   string
					id        string
				}{htmlBlock, content, id})
			}
		}

		return ast.WalkContinue, nil
	})

	// Process each details block by setting an attribute
	for _, item := range toProcess {
		// Store the ID as an attribute on the HTML block
		// We'll use this in the renderer
		item.htmlBlock.SetAttributeString("data-details-id", []byte(item.id))
	}
}

// generateID creates a unique ID for a details element based on its content
func (t *Transformer) generateID(content string) string {
	hash := sha256.Sum256([]byte(content))
	hashStr := hex.EncodeToString(hash[:])[:12]
	return fmt.Sprintf("%s%d-%s", t.idPrefix, t.counter, hashStr)
}

// NewTransformer creates a new Transformer
func NewTransformer(idPrefix string) *Transformer {
	return &Transformer{
		idPrefix: idPrefix,
		counter:  0,
	}
}
