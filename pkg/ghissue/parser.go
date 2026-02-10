package ghissue

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	internalRefRegexp = regexp.MustCompile(`^#([0-9]+)`)
)

type issueParser struct {
	config *Config
}

// NewParser creates a new inline parser for GitHub issue/PR references
func NewParser(config *Config) parser.InlineParser {
	return &issueParser{
		config: config,
	}
}

func (s *issueParser) Trigger() []byte {
	return []byte{'#'}
}

func (s *issueParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	// Don't parse inside code blocks or links
	if parent.Kind() == ast.KindCodeSpan || parent.Kind() == ast.KindLink {
		return nil
	}

	// Must start with #
	if len(line) == 0 || line[0] != '#' {
		return nil
	}

	// Check if this is part of an external reference (owner/repo#123)
	// by checking the preceding character
	preceding := block.PrecendingCharacter()
	if preceding != rune(text.EOF) && preceding != ' ' && preceding != '\t' && preceding != '\n' {
		// There's text immediately before the #, this might be part of owner/repo#123
		// Don't parse it here, let the transformer handle it
		return nil
	}

	// Parse as internal reference (#123)
	if m := internalRefRegexp.FindSubmatchIndex(line); m != nil {
		number := line[m[2]:m[3]]
		block.Advance(m[1])

		var repository []byte
		if s.config != nil && s.config.Repository != "" {
			repository = []byte(s.config.Repository)
		}

		return NewGitHubIssue(repository, number)
	}

	return nil
}
