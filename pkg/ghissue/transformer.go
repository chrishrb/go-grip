package ghissue

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	externalRefFullRegexp = regexp.MustCompile(`([a-zA-Z0-9][-a-zA-Z0-9]*)/([a-zA-Z0-9][-a-zA-Z0-9_]*)#([0-9]+)`)
	internalRefFullRegexp = regexp.MustCompile(`(?:^|[^a-zA-Z0-9/])#([0-9]+)`)
)

type transformer struct {
	config *Config
}

// NewTransformer creates a new AST transformer for GitHub issue/PR references
func NewTransformer(config *Config) parser.ASTTransformer {
	return &transformer{config: config}
}

func (t *transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// First pass: collect all text nodes that need processing
	type textNodeInfo struct {
		node   *ast.Text
		parent ast.Node
	}
	var textNodes []textNodeInfo

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Only process text nodes
		if n.Kind() != ast.KindText {
			return ast.WalkContinue, nil
		}

		parent := n.Parent()
		if parent == nil {
			return ast.WalkContinue, nil
		}

		// Don't process inside code or links
		if parent.Kind() == ast.KindCodeSpan || parent.Kind() == ast.KindLink {
			return ast.WalkContinue, nil
		}

		textNode := n.(*ast.Text)
		segment := textNode.Segment
		textBytes := segment.Value(reader.Source())

		// Check for both external and internal references
		extMatches := externalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1)
		intMatches := internalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1)

		if len(extMatches) == 0 && len(intMatches) == 0 {
			return ast.WalkContinue, nil
		}

		// Store this text node for processing
		textNodes = append(textNodes, textNodeInfo{
			node:   textNode,
			parent: parent,
		})

		return ast.WalkContinue, nil
	})

	// Second pass: process each text node
	for _, info := range textNodes {
		t.processTextNode(info.node, info.parent, reader)
	}
}

func (t *transformer) processTextNode(n *ast.Text, parent ast.Node, reader text.Reader) {
	segment := n.Segment
	textBytes := segment.Value(reader.Source())

	// Check if there's a previous sibling text node that might contain owner/repo
	var fullText []byte
	var precedingTextNode *ast.Text

	if prevSibling := n.PreviousSibling(); prevSibling != nil && prevSibling.Kind() == ast.KindText {
		prevText := prevSibling.(*ast.Text)
		prevBytes := prevText.Segment.Value(reader.Source())

		// Combine the two text nodes to check for external references
		fullText = append(append([]byte{}, prevBytes...), textBytes...)
		precedingTextNode = prevText
	} else {
		fullText = textBytes
	}

	// Check for external references in the combined text
	extMatches := externalRefFullRegexp.FindAllSubmatchIndex(fullText, -1)

	// If we found an external reference spanning both nodes, handle it carefully
	if len(extMatches) > 0 && precedingTextNode != nil {
		prevLen := len(fullText) - len(textBytes)
		for _, m := range extMatches {
			// If the match starts in the previous node and ends in current node
			if m[0] < prevLen && m[1] > prevLen {
				matchStart := m[0]

				// If there's text before the match in the preceding node, keep it
				if matchStart > 0 {
					prevSegment := precedingTextNode.Segment
					newSegment := text.NewSegment(
						prevSegment.Start,
						prevSegment.Start+matchStart,
					)
					precedingTextNode.Segment = newSegment
				} else {
					// No text before the match, remove the entire preceding node
					parent.RemoveChild(parent, precedingTextNode)
				}

				// Create the external GitHub issue node
				owner := fullText[m[2]:m[3]]
				repo := fullText[m[4]:m[5]]
				number := fullText[m[6]:m[7]]
				repository := append(append([]byte{}, owner...), '/')
				repository = append(repository, repo...)
				issueNode := NewExternalGitHubIssue(repository, number)

				// Replace current node with the issue node
				parent.ReplaceChild(parent, n, issueNode)
				return
			}
		}
	}

	// Otherwise, process normally with just the current text
	matches := t.findMatches(textBytes)
	if len(matches) == 0 {
		return
	}

	// Split the text node at each match and create issue nodes
	offset := 0
	for _, m := range matches {
		// Insert text before the match
		if m.start > offset {
			beforeText := ast.NewTextSegment(text.NewSegment(
				segment.Start+offset,
				segment.Start+m.start,
			))
			parent.InsertBefore(parent, n, beforeText)
		}

		// Insert the issue node
		issueNode := t.createIssueNode(m)
		parent.InsertBefore(parent, n, issueNode)
		offset = m.end
	}

	// Insert remaining text after the last match
	if offset < len(textBytes) {
		afterText := ast.NewTextSegment(text.NewSegment(
			segment.Start+offset,
			segment.Stop,
		))
		parent.InsertBefore(parent, n, afterText)
	}

	// Remove the original text node
	parent.RemoveChild(parent, n)
}

type issueMatch struct {
	start      int
	end        int
	isExternal bool
	owner      string
	repo       string
	number     string
}

func (t *transformer) findMatches(textBytes []byte) []issueMatch {
	var matches []issueMatch

	// Find external references (owner/repo#123)
	for _, m := range externalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1) {
		matches = append(matches, issueMatch{
			start:      m[0],
			end:        m[1],
			isExternal: true,
			owner:      string(textBytes[m[2]:m[3]]),
			repo:       string(textBytes[m[4]:m[5]]),
			number:     string(textBytes[m[6]:m[7]]),
		})
	}

	// Find internal references (#123)
	for _, m := range internalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1) {
		// Find the actual # position (regex may include preceding char)
		hashPos := m[0]
		for i := m[0]; i < m[2]; i++ {
			if textBytes[i] == '#' {
				hashPos = i
				break
			}
		}
		matches = append(matches, issueMatch{
			start:      hashPos,
			end:        m[1],
			isExternal: false,
			number:     string(textBytes[m[2]:m[3]]),
		})
	}

	// Sort matches by position (simple bubble sort for small arrays)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].start < matches[i].start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

func (t *transformer) createIssueNode(m issueMatch) *GitHubIssue {
	if m.isExternal {
		repo := m.owner + "/" + m.repo
		return NewExternalGitHubIssue([]byte(repo), []byte(m.number))
	}

	var repo []byte
	if t.config != nil && t.config.Repository != "" {
		repo = []byte(t.config.Repository)
	}
	return NewGitHubIssue(repo, []byte(m.number))
}
