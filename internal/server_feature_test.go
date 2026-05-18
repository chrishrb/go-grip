package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownResponseOmitsDisabledFeatureAssets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		disabledFeature string
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "mathjax",
			disabledFeature: FeatureMathJax,
			input:           "Inline math: $x + y$\n",
			wantContains:    []string{"$x + y$"},
			wantNotContains: []string{"/static/js/tex-mml-chtml.js", "/static/css/mathjax.css"},
		},
		{
			name:            "mermaid",
			disabledFeature: FeatureMermaid,
			input:           "```mermaid\ngraph TD;\n  A-->B;\n```\n",
			wantContains:    []string{`class="highlight`, `graph TD;`},
			wantNotContains: []string{"/static/js/mermaid.min.js", "/static/css/github-mermaid.css"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(tt.input), 0o644); err != nil {
				t.Fatalf("write README.md: %v", err)
			}

			server := NewServer("localhost", 6419, false, false, false, NewParser([]string{tt.disabledFeature}))
			handler := server.newHandler(http.Dir(tmpDir))

			req := httptest.NewRequest(http.MethodGet, "/README.md", nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			body := recorder.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Fatalf("expected body to contain %q, got %q", want, body)
				}
			}
			for _, unwanted := range tt.wantNotContains {
				if strings.Contains(body, unwanted) {
					t.Fatalf("expected body not to contain %q, got %q", unwanted, body)
				}
			}
		})
	}
}
