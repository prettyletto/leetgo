package tui

import (
	"os"
	"testing"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteNeovimDebugBootstrap(t *testing.T) {
	spec, ok := generator.SpecForProblem(&roadmap.Problem{ID: 424, Slug: "longest-repeating-character-replacement"})
	require.True(t, ok)
	dir := t.TempDir()

	path, err := writeNeovimDebugBootstrap(dir, spec)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `require("dap")`)
	assert.Contains(t, content, `require, "dapui"`)
	assert.Contains(t, content, `nvim-dap-virtual-text`)
	assert.Contains(t, content, `command = "dlv"`)
	assert.Contains(t, content, `args = { "dap", "-l", "127.0.0.1:${port}" }`)
	assert.Contains(t, content, `"-test.run", "^TestLeetgoDebugCase$"`)
	assert.Contains(t, content, `<S-F11>`)
	assert.Contains(t, content, `<leader>dr`)
	assert.Contains(t, content, `<leader>dx`)
	assert.Contains(t, content, `dap.set_breakpoint()`)
}

func TestNeovimFuncSearchCommand(t *testing.T) {
	spec, ok := generator.SpecForProblem(&roadmap.Problem{ID: 424, Slug: "longest-repeating-character-replacement"})
	require.True(t, ok)

	assert.Equal(t, "+/func characterReplacement", neovimFuncSearchCommand(spec))
}
