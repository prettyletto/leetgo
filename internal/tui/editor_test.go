package tui

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditorCommand_WrapsTerminalEditor(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	cmd := editorCommand("nvim", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "xdg-terminal-exec", filepath.Base(cmd.Path))
	assert.Equal(t, "/bin/zsh", cmd.Args[1])
	assert.Equal(t, "-i", cmd.Args[2])
	assert.Equal(t, "-c", cmd.Args[3])
	assert.Equal(t, "nvim two_sum.go two_sum_test.go", cmd.Args[4])
	assert.Equal(t, "/tmp/problem", cmd.Dir)
}

func TestEditorCommand_ShellEscapesSpacesInPaths(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	cmd := editorCommand("nvim", []string{"/tmp/my workspace/two_sum.go"}, "/tmp/problem")

	assert.Equal(t, "xdg-terminal-exec", filepath.Base(cmd.Path))
	assert.Equal(t, "nvim '/tmp/my workspace/two_sum.go'", cmd.Args[4])
}

func TestEditorCommand_UsesGraphicalEditorDirectly(t *testing.T) {
	t.Setenv("TERMINAL", "xdg-terminal-exec")

	cmd := editorCommand("code --reuse-window", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "code", cmd.Path)
	assert.Equal(t, []string{"code", "--reuse-window", "two_sum.go", "two_sum_test.go"}, cmd.Args)
	assert.Equal(t, "/tmp/problem", cmd.Dir)
}

func TestEditorCommand_FallsBackToXDGOpen(t *testing.T) {
	cmd := editorCommand("", []string{"two_sum.go", "two_sum_test.go"}, "/tmp/problem")

	assert.Equal(t, "xdg-open", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"xdg-open", "/tmp/problem"}, cmd.Args)
	assert.Equal(t, "/tmp/problem", cmd.Dir)
}
