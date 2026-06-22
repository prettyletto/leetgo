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

func TestWriteAndReadManifest(t *testing.T) {
	dir := t.TempDir()

	m := &Manifest{
		ProblemID:     1,
		Slug:          "two-sum",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "go",
		StubPath:      "two_sum.go",
		TestsuitePath: "two_sum_test.go",
	}
	require.NoError(t, WriteManifest(dir, m))

	assert.FileExists(t, filepath.Join(dir, ManifestFileName))

	read, foundDir, err := ReadManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, dir, foundDir)
	assert.Equal(t, m.ProblemID, read.ProblemID)
	assert.Equal(t, m.Slug, read.Slug)
	assert.Equal(t, m.Roadmap, read.Roadmap)
	assert.Equal(t, m.Language, read.Language)
	assert.Equal(t, m.StubPath, read.StubPath)
	assert.Equal(t, m.TestsuitePath, read.TestsuitePath)
}

func TestWriteManifest_SameProblemID_Overwrites(t *testing.T) {
	dir := t.TempDir()

	m := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "two_sum.go", TestsuitePath: "two_sum_test.go"}
	require.NoError(t, WriteManifest(dir, m))

	updated := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "python", StubPath: "two_sum.py", TestsuitePath: "two_sum_test.py"}
	require.NoError(t, WriteManifest(dir, updated))

	read, _, err := ReadManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, "python", read.Language)
	assert.Equal(t, "two_sum.py", read.StubPath)
}

func TestWriteManifest_DifferentProblemID_Error(t *testing.T) {
	dir := t.TempDir()

	m := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "two_sum.go", TestsuitePath: "two_sum_test.go"}
	require.NoError(t, WriteManifest(dir, m))

	other := &Manifest{ProblemID: 49, Slug: "group-anagrams", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "group_anagrams.go", TestsuitePath: "group_anagrams_test.go"}
	err := WriteManifest(dir, other)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "different Problem ID")
}

func TestEnsureManifestWritable_IgnoresParentManifest(t *testing.T) {
	dir := t.TempDir()
	parent := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "two_sum.go", TestsuitePath: "two_sum_test.go"}
	require.NoError(t, WriteManifest(dir, parent))

	child := filepath.Join(dir, "child")
	require.NoError(t, os.MkdirAll(child, 0o755))

	require.NoError(t, EnsureManifestWritable(child, 49))
}

func TestEnsureManifestWritable_DifferentProblemID_Error(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "two_sum.go", TestsuitePath: "two_sum_test.go"}
	require.NoError(t, WriteManifest(dir, m))

	err := EnsureManifestWritable(dir, 49)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "different Problem ID")
}

func TestReadManifest_ParentWalking(t *testing.T) {
	dir := t.TempDir()

	m := &Manifest{ProblemID: 1, Slug: "two-sum", Roadmap: "from-zero-to-hero", Stage: "arrays-hashing", Language: "go", StubPath: "two_sum.go", TestsuitePath: "two_sum_test.go"}
	require.NoError(t, WriteManifest(dir, m))

	subdir := filepath.Join(dir, "sub", "deep")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	read, foundDir, err := ReadManifest(subdir)
	require.NoError(t, err)
	assert.Equal(t, dir, foundDir)
	assert.Equal(t, 1, read.ProblemID)
}

func TestReadManifest_NotFound(t *testing.T) {
	dir := t.TempDir()

	read, foundDir, err := ReadManifest(dir)
	require.NoError(t, err)
	assert.Nil(t, read)
	assert.Empty(t, foundDir)
}

func TestManager_Generate_WritesManifest(t *testing.T) {
	root := t.TempDir()
	gen := generator.New()
	mgr := New(root, gen)

	p := &roadmap.Problem{
		ID:       1,
		Title:    "Two Sum",
		Slug:     "two-sum",
		Category: "arrays-hashing",
		Stage:    "arrays-hashing",
	}

	stubPath, testPath, err := mgr.Generate(p, generator.LangGo)
	require.NoError(t, err)

	dir := mgr.ProblemDir(p)
	m := &Manifest{
		ProblemID:     p.ID,
		Slug:          p.Slug,
		Roadmap:       "from-zero-to-hero",
		Stage:         p.Stage,
		Language:      "go",
		StubPath:      filepath.Base(stubPath),
		TestsuitePath: filepath.Base(testPath),
	}
	require.NoError(t, WriteManifest(dir, m))

	assert.FileExists(t, filepath.Join(dir, ManifestFileName))
	read, _, err := ReadManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, p.ID, read.ProblemID)
	assert.Equal(t, "go", read.Language)
}
