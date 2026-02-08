package alert

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var alertRegex = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*`)

// Transformer is a transformer that converts blockquotes with alert syntax to Alert nodes
type Transformer struct{}

// Transform transforms blockquotes into Alert nodes if they match the GitHub alert syntax
func (t *Transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	var toTransform []struct {
		blockquote *ast.Blockquote
		alertType  AlertType
	}

	// First pass: identify blockquotes to transform
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		blockquote, ok := n.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}

		firstChild := blockquote.FirstChild()
		if firstChild == nil || firstChild.Kind() != ast.KindParagraph {
			return ast.WalkContinue, nil
		}

		paragraph := firstChild.(*ast.Paragraph)
		
		// Collect all text from the first line (may be split across multiple nodes)
		var firstLineText strings.Builder
		var textNodes []*ast.Text
		
		for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
			if textNode, ok := child.(*ast.Text); ok {
				firstLineText.Write(textNode.Segment.Value(source))
				textNodes = append(textNodes, textNode)
				// Check if this is a hard line break
				if textNode.HardLineBreak() || textNode.SoftLineBreak() {
					break
				}
			} else {
				// Stop at first non-text node
				break
			}
		}
		
		if len(textNodes) == 0 {
			return ast.WalkContinue, nil
		}
		
		firstText := firstLineText.String()
		matches := alertRegex.FindStringSubmatch(firstText)
		if matches == nil {
			return ast.WalkContinue, nil
		}

		alertTypeStr := strings.ToLower(matches[1])
		toTransform = append(toTransform, struct {
			blockquote *ast.Blockquote
			alertType  AlertType
		}{
			blockquote: blockquote,
			alertType:  AlertType(alertTypeStr),
		})

		return ast.WalkContinue, nil
	})

	// Second pass: transform the identified blockquotes
	for _, item := range toTransform {
		blockquote := item.blockquote
		alertType := item.alertType

		// Create the Alert node
		alertNode := NewAlert(alertType)

		// Copy attributes
		for _, attr := range blockquote.Attributes() {
			alertNode.SetAttribute(attr.Name, attr.Value)
		}

		// Remove the [!TYPE] marker from the first line
		firstChild := blockquote.FirstChild()
		if firstChild != nil && firstChild.Kind() == ast.KindParagraph {
			paragraph := firstChild.(*ast.Paragraph)
			
			// Collect all text nodes from the first line
			var textNodes []*ast.Text
			var combinedText strings.Builder
			
			for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
				if textNode, ok := child.(*ast.Text); ok {
					combinedText.Write(textNode.Segment.Value(source))
					textNodes = append(textNodes, textNode)
					if textNode.HardLineBreak() || textNode.SoftLineBreak() {
						break
					}
				} else {
					break
				}
			}
			
			// Find where the marker ends in the combined text
			markerMatch := alertRegex.FindStringIndex(combinedText.String())
			if markerMatch != nil {
				markerEndPos := markerMatch[1]
				
				// Now figure out which nodes to remove/modify
				var currentPos int
				for _, textNode := range textNodes {
					nodeLen := len(textNode.Segment.Value(source))
					nodeEndPos := currentPos + nodeLen
					
					if markerEndPos <= currentPos {
						// This node is after the marker, keep it
						break
					} else if markerEndPos >= nodeEndPos {
						// This entire node is part of the marker, remove it
						paragraph.RemoveChild(paragraph, textNode)
					} else {
						// Marker ends in the middle of this node
						offsetInNode := markerEndPos - currentPos
						oldSegment := textNode.Segment
						newStart := oldSegment.Start + offsetInNode
						if newStart >= oldSegment.Stop {
							paragraph.RemoveChild(paragraph, textNode)
						} else {
							textNode.Segment = text.NewSegment(newStart, oldSegment.Stop)
						}
						break
					}
					
					currentPos = nodeEndPos
				}
			}
			
			// If paragraph is now empty, remove it
			if paragraph.ChildCount() == 0 {
				blockquote.RemoveChild(blockquote, paragraph)
			}
		}

		// Move all children from blockquote to alert
		for child := blockquote.FirstChild(); child != nil; {
			next := child.NextSibling()
			blockquote.RemoveChild(blockquote, child)
			alertNode.AppendChild(alertNode, child)
			child = next
		}

		// Replace blockquote with alert
		parent := blockquote.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, blockquote, alertNode)
		}
	}
}

// NewTransformer creates a new Transformer
func NewTransformer() *Transformer {
	return &Transformer{}
}
