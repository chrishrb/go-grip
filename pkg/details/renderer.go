package details

import (
	"fmt"
	"strings"
	"sync"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// HTMLRenderer renders the state management script for details elements
type HTMLRenderer struct {
	html.Config
	mu             sync.Mutex
	scriptRendered bool
}

// NewHTMLRenderer creates a new HTMLRenderer
func NewHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &HTMLRenderer{
		Config:         html.NewConfig(),
		scriptRendered: false,
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers rendering functions
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Register for HTMLBlock to inject IDs
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	// Register for Document nodes to inject script at the end
	reg.Register(ast.KindDocument, r.renderDocument)
}

// renderHTMLBlock renders HTML blocks, injecting IDs for details elements
func (r *HTMLRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	htmlBlock := node.(*ast.HTMLBlock)

	// Check if this block has a details ID attribute
	if attr, ok := htmlBlock.Attribute([]byte("data-details-id")); ok {
		if idBytes, ok := attr.([]byte); ok {
			id := string(idBytes)

			// Get the HTML content
			var htmlContent strings.Builder
			for i := 0; i < htmlBlock.Lines().Len(); i++ {
				line := htmlBlock.Lines().At(i)
				htmlContent.Write(line.Value(source))
			}

			content := htmlContent.String()
			modified := injectID(content, id)
			_, _ = w.WriteString(modified)
			return ast.WalkContinue, nil
		}
	}

	// Default rendering for non-details HTML blocks
	if htmlBlock.HasClosure() {
		_, _ = w.WriteString("<!-- raw HTML omitted -->\n")
		return ast.WalkContinue, nil
	}

	l := htmlBlock.Lines().Len()
	for i := range l {
		line := htmlBlock.Lines().At(i)
		_, _ = w.Write(line.Value(source))
	}

	return ast.WalkContinue, nil
}

// renderDocument renders the state management script once per document
func (r *HTMLRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		// At the end of the document, inject the script if we haven't already
		r.mu.Lock()
		defer r.mu.Unlock()

		if !r.scriptRendered {
			// Check if document contains any details elements
			hasDetails := false
			_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering {
					return ast.WalkContinue, nil
				}
				if htmlBlock, ok := n.(*ast.HTMLBlock); ok {
					if _, ok := htmlBlock.Attribute([]byte("data-details-id")); ok {
						hasDetails = true
						return ast.WalkStop, nil
					}
				}
				return ast.WalkContinue, nil
			})

			if hasDetails {
				_, _ = w.WriteString("\n")
				_, _ = w.WriteString(getStateManagementScript())
				r.scriptRendered = true
			}
		}
	}
	return ast.WalkContinue, nil
}

// injectID adds an ID attribute to the <details> tag
func injectID(html string, id string) string {
	html = strings.TrimSpace(html)

	// Find the end of the opening <details> tag
	detailsEnd := strings.Index(html, ">")
	if detailsEnd == -1 {
		return html
	}

	// Check if it already has an id attribute
	if strings.Contains(html[:detailsEnd], " id=") {
		return html
	}

	// Check if it's a self-closing tag (shouldn't be for details)
	beforeClose := strings.TrimSpace(html[:detailsEnd])
	if strings.HasSuffix(beforeClose, "/") {
		return beforeClose[:len(beforeClose)-1] + fmt.Sprintf(` id="%s"`, id) + " />" + html[detailsEnd+1:]
	}

	// Insert ID before the closing >
	return html[:detailsEnd] + fmt.Sprintf(` id="%s"`, id) + html[detailsEnd:]
}

// getStateManagementScript returns the JavaScript code for state management
func getStateManagementScript() string {
	return `<script>
(function() {
  'use strict';
  
  // Initialize state management for all details elements
  function initDetailsState() {
    const details = document.querySelectorAll('details[id]');
    
    details.forEach(function(detail) {
      const id = detail.id;
      
      // Restore state from session storage
      const savedState = sessionStorage.getItem('details-state-' + id);
      if (savedState === 'open') {
        detail.open = true;
      } else if (savedState === 'closed') {
        detail.open = false;
      }
      
      // Save state on toggle
      detail.addEventListener('toggle', function() {
        const state = detail.open ? 'open' : 'closed';
        sessionStorage.setItem('details-state-' + id, state);
      });
    });
  }
  
  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initDetailsState);
  } else {
    initDetailsState();
  }
})();
</script>
`
}
