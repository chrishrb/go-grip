package defaults

import (
	"os/exec"
	"testing"
)

func TestThemeSwitchPictureSources(t *testing.T) {
	t.Parallel()

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found; skipping theme-switch.js tests")
	}
	out, err := exec.Command(node, "testdata/theme-switch.test.js").CombinedOutput()
	if err != nil {
		t.Fatalf("theme-switch.js tests failed: %v\n%s", err, out)
	}
}
