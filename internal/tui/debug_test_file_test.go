package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteGoDebugTestFile(t *testing.T) {
	spec, ok := generator.SpecForProblem(&roadmap.Problem{ID: 1, Slug: "two-sum"})
	require.True(t, ok)
	dir := t.TempDir()

	path, err := writeGoDebugTestFile(dir, spec, 1)
	require.NoError(t, err)
	assert.Equal(t, "leetgo_debug_test.go", filepath.Base(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "func TestLeetgoDebugCase(t *testing.T)")
	assert.Contains(t, content, "nums := []int{3,2,4}")
	assert.Contains(t, content, "target := 6")
	assert.Contains(t, content, "got := twoSum(nums, target)")
	assert.Contains(t, content, "expect := []int{1,2}")
}
