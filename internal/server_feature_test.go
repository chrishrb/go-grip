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

			server := NewServer("localhost", 6419, false, false, false, NewParser([]string{tt.disabledFeature}), nil)
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

func TestResponsesNameMermaidIconPacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		iconPacks []string
		want      string
	}{
		{
			name: "none",
			want: `<script src="/static/js/mermaid-init.js"></script>`,
		},
		{
			name:      "two packs",
			iconPacks: []string{"logos", "fa6-brands"},
			want:      `<script src="/static/js/mermaid-init.js" data-icon-packs="logos,fa6-brands"></script>`,
		},
		{
			name:      "spaces and capitals",
			iconPacks: []string{" Logos "},
			want:      `<script src="/static/js/mermaid-init.js" data-icon-packs="logos"></script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := NewServer("localhost", 6419, false, false, false, NewParser(nil), nil)
			if err := server.SetMermaidIconPacks(tt.iconPacks); err != nil {
				t.Fatalf("SetMermaidIconPacks(%q): %v", tt.iconPacks, err)
			}
			handler := server.newHandler(http.Dir(writeTree(t, map[string]string{"README.md": "# Title\n"})))

			for _, path := range []string{"/README.md", "/"} {
				req := httptest.NewRequest(http.MethodGet, path, nil)
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)

				if body := recorder.Body.String(); !strings.Contains(body, tt.want) {
					t.Errorf("GET %s: expected body to contain %q, got %q", path, tt.want, body)
				}
			}
		})
	}
}

func TestSetMermaidIconPacksRejectsNonPrefixNames(t *testing.T) {
	t.Parallel()

	server := NewServer("localhost", 6419, false, false, false, NewParser(nil), nil)
	for _, name := range []string{"", "-logos", "logos-", "simple--icons", "logos,mdi", `logos" onload="alert(1)`} {
		if err := server.SetMermaidIconPacks([]string{name}); err == nil {
			t.Errorf("SetMermaidIconPacks(%q) = nil, want an error", name)
		}
	}
}
