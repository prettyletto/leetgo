package tui

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditorCommand_WrapsTerminalEditor(t *testing.T) {
	t.Setenv("TERMINAL", "xdg-terminal-exec")

	cmd := editorCommand("nvim", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "xdg-terminal-exec", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"xdg-terminal-exec", "nvim", "two_sum.go", "two_sum_test.go"}, cmd.Args)
}

func TestEditorCommand_UsesGraphicalEditorDirectly(t *testing.T) {
	t.Setenv("TERMINAL", "xdg-terminal-exec")

	cmd := editorCommand("code --reuse-window", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "code", cmd.Path)
	assert.Equal(t, []string{"code", "--reuse-window", "two_sum.go", "two_sum_test.go"}, cmd.Args)
}

func TestEditorCommand_FallsBackToXDGOpen(t *testing.T) {
	cmd := editorCommand("", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "xdg-open", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"xdg-open", "/tmp/problem"}, cmd.Args)
}
