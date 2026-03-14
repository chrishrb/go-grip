package mermaid

import "strings"

// unlines returns the string formed by joining the provided strings after
// appending a newline to each.
func unlines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
