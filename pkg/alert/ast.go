package alert

import (
	"github.com/yuin/goldmark/ast"
)

// AlertType represents the type of alert (note, tip, important, warning, caution)
type AlertType string

const (
	AlertTypeNote      AlertType = "note"
	AlertTypeTip       AlertType = "tip"
	AlertTypeImportant AlertType = "important"
	AlertTypeWarning   AlertType = "warning"
	AlertTypeCaution   AlertType = "caution"
)

// KindAlert is the kind for Alert nodes
var KindAlert = ast.NewNodeKind("Alert")

// Alert represents a GitHub-style alert block
type Alert struct {
	ast.BaseBlock
	AlertType AlertType
}

// NewAlert creates a new Alert node
func NewAlert(alertType AlertType) *Alert {
	return &Alert{
		BaseBlock: ast.BaseBlock{},
		AlertType: alertType,
	}
}

// Kind returns the kind of this node
func (n *Alert) Kind() ast.NodeKind {
	return KindAlert
}

// Dump dumps the alert node
func (n *Alert) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Type": string(n.AlertType),
	}, nil)
}
