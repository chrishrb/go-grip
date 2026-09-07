// Package slug generates heading anchor ids the way GitHub does.
//
// Goldmark's built-in generator turns underscores into dashes and drops
// non-ASCII characters, so links like `#implementation-attack_pattern` written
// against GitHub's ids don't resolve. GitHub keeps word characters (letters,
// digits, marks and the underscore) plus the dash, lowercases everything,
// turns spaces into dashes and drops the rest.
package slug

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// IDs implements parser.IDs with GitHub's slug rules.
type IDs struct {
	values map[string]bool
}

// NewIDs returns a parser.IDs generating GitHub-compatible anchor ids.
func NewIDs() parser.IDs {
	return &IDs{values: map[string]bool{}}
}

// Generate generates a new element id for the given value.
func (s *IDs) Generate(value []byte, kind ast.NodeKind) []byte {
	result := []byte(Slug(string(value)))
	if len(result) == 0 {
		if kind == ast.KindHeading {
			result = []byte("heading")
		} else {
			result = []byte("id")
		}
	}

	if !s.values[string(result)] {
		s.values[string(result)] = true
		return result
	}
	for i := 1; ; i++ {
		newResult := fmt.Sprintf("%s-%d", result, i)
		if !s.values[newResult] {
			s.values[newResult] = true
			return []byte(newResult)
		}
	}
}

// Put puts a given element id to the used ids table.
func (s *IDs) Put(value []byte) {
	s.values[string(value)] = true
}

// Slug converts a heading text into a GitHub-compatible anchor id.
func Slug(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r):
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}
