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
	return &transformer{
		config: config,
	}
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

	// Check for both external and internal references in the combined text
	extMatches := externalRefFullRegexp.FindAllSubmatchIndex(fullText, -1)

	// If we found an external reference in the combined text spanning both nodes,
	// we need to handle it carefully to preserve text before the owner/repo pattern
	if len(extMatches) > 0 && precedingTextNode != nil {
		// Check if the external reference spans across the two nodes
		prevLen := len(fullText) - len(textBytes)
		for _, m := range extMatches {
			// If the match starts in the previous node and ends in current node
			if m[0] < prevLen && m[1] > prevLen {
				// Calculate where the owner/repo pattern starts
				matchStart := m[0]

				// If there's text before the match in the preceding node, keep it
				if matchStart > 0 {
					// Update the preceding text node to only contain text before the match
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
	extMatches = externalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1)
	intMatches := internalRefFullRegexp.FindAllSubmatchIndex(textBytes, -1)

	// Combine and sort all matches
	type match struct {
		start      int
		end        int
		isExternal bool
		owner      []byte
		repo       []byte
		number     []byte
	}

	var allMatches []match

	for _, m := range extMatches {
		allMatches = append(allMatches, match{
			start:      m[0],
			end:        m[1],
			isExternal: true,
			owner:      textBytes[m[2]:m[3]],
			repo:       textBytes[m[4]:m[5]],
			number:     textBytes[m[6]:m[7]],
		})
	}

	for _, m := range intMatches {
		// For internal matches, the regex includes a preceding char if not at start
		numStart := m[2]
		numEnd := m[3]
		// Find where the # actually starts
		hashPos := m[0]
		for i := m[0]; i < numStart; i++ {
			if textBytes[i] == '#' {
				hashPos = i
				break
			}
		}
		allMatches = append(allMatches, match{
			start:      hashPos,
			end:        m[1],
			isExternal: false,
			number:     textBytes[numStart:numEnd],
		})
	}

	// Sort matches by position
	for i := 0; i < len(allMatches)-1; i++ {
		for j := i + 1; j < len(allMatches); j++ {
			if allMatches[j].start < allMatches[i].start {
				allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
			}
		}
	}

	// Process each match and split the text node
	offset := 0
	for _, m := range allMatches {
		// Create text node for content before the match
		if m.start > offset {
			beforeSegment := text.NewSegment(
				segment.Start+offset,
				segment.Start+m.start,
			)
			beforeText := ast.NewTextSegment(beforeSegment)
			parent.InsertBefore(parent, n, beforeText)
		}

		// Create GitHub issue node
		var issueNode *GitHubIssue
		if m.isExternal {
			repository := append(append([]byte{}, m.owner...), '/')
			repository = append(repository, m.repo...)
			issueNode = NewExternalGitHubIssue(repository, m.number)
		} else {
			var repository []byte
			if t.config != nil && t.config.Repository != "" {
				repository = []byte(t.config.Repository)
			}
			issueNode = NewGitHubIssue(repository, m.number)
		}

		parent.InsertBefore(parent, n, issueNode)

		offset = m.end
	}

	// Create text node for content after the last match
	if offset < len(textBytes) {
		afterSegment := text.NewSegment(
			segment.Start+offset,
			segment.Stop,
		)
		afterText := ast.NewTextSegment(afterSegment)
		parent.InsertBefore(parent, n, afterText)
	}

	// Remove the original text node
	parent.RemoveChild(parent, n)
}
