package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Generate_Go(t *testing.T) {
	root := t.TempDir()
	gen := generator.New()
	m := New(root, gen)

	p := &roadmap.Problem{
		ID:       1,
		Title:    "Two Sum",
		Slug:     "two-sum",
		Category: "arrays-hashing",
	}

	stubPath, testPath, err := m.Generate(p, generator.LangGo)
	require.NoError(t, err)

	assert.FileExists(t, stubPath)
	assert.FileExists(t, testPath)
	assert.Contains(t, stubPath, "arrays-hashing/1-two-sum/two_sum.go")
	assert.Contains(t, testPath, "arrays-hashing/1-two-sum/two_sum_test.go")

	stub, err := os.ReadFile(stubPath)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "func TwoSum")
}

func TestManager_Generate_Python(t *testing.T) {
	root := t.TempDir()
	gen := generator.New()
	m := New(root, gen)

	p := &roadmap.Problem{
		ID:       1,
		Title:    "Two Sum",
		Slug:     "two-sum",
		Category: "arrays-hashing",
	}

	stubPath, testPath, err := m.Generate(p, generator.LangPython)
	require.NoError(t, err)

	assert.FileExists(t, stubPath)
	assert.FileExists(t, testPath)
	assert.Contains(t, stubPath, "two_sum.py")
	assert.Contains(t, testPath, "two_sum_test.py")
}

func TestManager_Generate_UnsupportedLanguage(t *testing.T) {
	root := t.TempDir()
	gen := generator.New()
	m := New(root, gen)

	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}

	_, _, err := m.Generate(p, "ruby")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported language")
}

func TestManager_ProblemDir(t *testing.T) {
	m := New("/workspace", generator.New())
	p := &roadmap.Problem{ID: 49, Slug: "group-anagrams", Category: "arrays-hashing"}

	dir := m.ProblemDir(p)
	assert.Equal(t, filepath.Join("/workspace", "arrays-hashing", "49-group-anagrams"), dir)
}
