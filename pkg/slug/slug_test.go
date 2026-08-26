package slug

import (
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"underscores are kept", "Implementation `ATTACK_PATTERN`", "implementation-attack_pattern"},
		{"spaces become dashes", "Hello World", "hello-world"},
		{"punctuation is dropped", "What's new? (v2.0)", "whats-new-v20"},
		{"dashes are kept", "go-grip is great", "go-grip-is-great"},
		{"non-ascii letters are kept", "Überschrift mit Ümlaut", "überschrift-mit-ümlaut"},
		{"cjk is kept", "日本語の見出し", "日本語の見出し"},
		// GitHub drops the emoji but keeps both surrounding spaces as dashes
		{"emoji are dropped", "Release 🎉 notes", "release--notes"},
		{"surrounding space is trimmed", "  spaced out  ", "spaced-out"},
		{"only punctuation is empty", "!!!", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slug(tt.input); got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIDsGenerateDeduplicates(t *testing.T) {
	ids := NewIDs()
	for i, want := range []string{"heading", "heading-1", "heading-2"} {
		got := string(ids.Generate([]byte("Heading"), ast.KindHeading))
		if got != want {
			t.Errorf("generation %d = %q, want %q", i, got, want)
		}
	}
}

func TestIDsGenerateFallback(t *testing.T) {
	ids := NewIDs()
	if got := string(ids.Generate([]byte("###"), ast.KindHeading)); got != "heading" {
		t.Errorf("heading fallback = %q, want %q", got, "heading")
	}
	if got := string(ids.Generate([]byte("###"), ast.KindLink)); got != "id" {
		t.Errorf("non-heading fallback = %q, want %q", got, "id")
	}
}

func TestIDsPut(t *testing.T) {
	ids := NewIDs()
	ids.Put([]byte("taken"))
	if got := string(ids.Generate([]byte("Taken"), ast.KindHeading)); got != "taken-1" {
		t.Errorf("got %q, want %q", got, "taken-1")
	}
}
