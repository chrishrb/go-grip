package ghissue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

func TestHTMLRendererIntegration(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		input        string
		wantContains []string
		wantClass    string
	}{
		{
			name:         "internal issue with repository",
			repo:         "owner/repo",
			input:        "See #123 for details",
			wantContains: []string{"https://github.com/owner/repo/", "#123", "issue-link"},
			wantClass:    "issue-link",
		},
		{
			name:         "external issue reference",
			repo:         "",
			input:        "Check out grafana/grafana#10",
			wantContains: []string{"https://github.com/grafana/grafana/", "grafana/grafana#10"},
		},
		{
			name:         "multiple internal references",
			repo:         "test/project",
			input:        "See #100 and #200",
			wantContains: []string{"#100", "#200", "https://github.com/test/project/"},
		},
		{
			name:         "multiple external references same repo",
			repo:         "",
			input:        "Check grafana/grafana#10 and grafana/grafana#20",
			wantContains: []string{"grafana/grafana#10", "grafana/grafana#20", "https://github.com/grafana/grafana/"},
		},
		{
			name:         "multiple external references different repos",
			repo:         "",
			input:        "See kubernetes/kubernetes#100 and docker/docker#200",
			wantContains: []string{"kubernetes/kubernetes#100", "docker/docker#200", "https://github.com/kubernetes/kubernetes/", "https://github.com/docker/docker/"},
		},
		{
			name:         "mixed internal and external references",
			repo:         "myorg/myrepo",
			input:        "Related to #50, see also grafana/grafana#100 and #60",
			wantContains: []string{"#50", "grafana/grafana#100", "#60", "https://github.com/myorg/myrepo/", "https://github.com/grafana/grafana/"},
		},
		{
			name:         "three issues in a row",
			repo:         "owner/repo",
			input:        "Issues: #1, #2, #3",
			wantContains: []string{"#1", "#2", "#3"},
		},
		{
			name:         "external refs without spaces",
			repo:         "",
			input:        "org1/repo1#5,org2/repo2#10",
			wantContains: []string{"org1/repo1#5", "org2/repo2#10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(WithRepository(tt.repo)),
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestHTMLRendererRegistration(t *testing.T) {
	config := &Config{
		Repository: "test/repo",
	}

	r := NewHTMLRenderer(config)

	// Create a mock registerer to verify the registration
	registeredKinds := make(map[interface{}]bool)
	mockReg := &mockNodeRendererFuncRegisterer{
		registered: registeredKinds,
	}

	r.RegisterFuncs(mockReg)

	if !registeredKinds[KindGitHubIssue] {
		t.Error("Expected KindGitHubIssue to be registered")
	}
}

// Mock implementation of NodeRendererFuncRegisterer for testing
type mockNodeRendererFuncRegisterer struct {
	registered map[interface{}]bool
}

func (m *mockNodeRendererFuncRegisterer) Register(kind ast.NodeKind, fn renderer.NodeRendererFunc) {
	m.registered[kind] = true
}

func TestRendererNilConfig(t *testing.T) {
	// Should not panic with nil config
	r := NewHTMLRenderer(nil)
	if r == nil {
		t.Fatal("Expected non-nil renderer")
	}
}

func TestGitHubIssueNode(t *testing.T) {
	// Test node creation and properties
	repo := []byte("owner/repo")
	number := []byte("123")

	node := NewGitHubIssue(repo, number)

	if node.Kind() != KindGitHubIssue {
		t.Errorf("Expected kind %v, got %v", KindGitHubIssue, node.Kind())
	}

	if string(node.Repository) != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got %s", node.Repository)
	}

	if string(node.Number) != "123" {
		t.Errorf("Expected number '123', got %s", node.Number)
	}
}

func TestGitHubIssueNodeDump(t *testing.T) {
	// Test that Dump doesn't panic
	node := NewGitHubIssue([]byte("owner/repo"), []byte("123"))
	source := []byte("test")

	// Should not panic
	node.Dump(source, 0)
}

func TestRenderWithoutRepository(t *testing.T) {
	// Test rendering when no repository is configured
	md := goldmark.New(
		goldmark.WithExtensions(
			New(), // No repository configured
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	input := "See #123"
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	// Without repository, it might render as plain text or not match
	// Just verify it doesn't crash
	if output == "" {
		t.Error("Expected some output")
	}
}

func TestParserWithSource(t *testing.T) {
	// Test the parser directly with text source
	config := &Config{Repository: "test/repo"}
	p := NewParser(config)

	source := []byte("#123")
	reader := text.NewReader(source)
	ctx := parser.NewContext()

	// Create a simple parent node
	parent := ast.NewDocument()

	node := p.Parse(parent, reader, ctx)
	if node == nil {
		t.Error("Expected parser to return a node for #123")
	}

	if node != nil {
		if node.Kind() != KindGitHubIssue {
			t.Errorf("Expected GitHubIssue node, got %v", node.Kind())
		}
	}
}

func TestParserExternalReference(t *testing.T) {
	// External references are handled by the transformer, not the inline parser
	// Test that the transformer is working via integration test
	md := goldmark.New(
		goldmark.WithExtensions(
			New(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	input := "grafana/grafana#10"
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "grafana/grafana") {
		t.Errorf("Expected output to contain 'grafana/grafana', got: %s", output)
	}
}
