package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/prettyletto/leetgo/internal/catalog"
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
	assert.Contains(t, string(stub), "func twoSum")
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

func TestManager_Generate_WritesConcreteTwoSumFilesForAllLanguages(t *testing.T) {
	tests := []struct {
		name         string
		language     generator.Language
		stubFile     string
		testFile     string
		stubContains []string
		testContains []string
	}{
		{
			name:     "go",
			language: generator.LangGo,
			stubFile: "two_sum.go",
			testFile: "two_sum_test.go",
			stubContains: []string{
				"package solution",
				"func twoSum(nums []int, target int) []int",
				"return nil",
			},
			testContains: []string{
				"TestTwoSum",
				"[]int{2,7,11,15}",
				"[]int{0,1}",
			},
		},
		{
			name:     "python",
			language: generator.LangPython,
			stubFile: "two_sum.py",
			testFile: "two_sum_test.py",
			stubContains: []string{
				"def twoSum(nums: list[int], target: int) -> list[int]:",
			},
			testContains: []string{
				"from two_sum import twoSum",
				"twoSum",
			},
		},
		{
			name:     "typescript",
			language: generator.LangTypeScript,
			stubFile: "two_sum.ts",
			testFile: "two_sum.test.ts",
			stubContains: []string{
				"export function twoSum(nums: number[], target: number): number[]",
			},
			testContains: []string{
				"import { twoSum } from './two_sum'",
				"twoSum(",
			},
		},
		{
			name:     "java",
			language: generator.LangJava,
			stubFile: "two_sum.java",
			testFile: "two_sumTest.java",
			stubContains: []string{
				"class Solution",
				"public int[] twoSum",
			},
			testContains: []string{
				"Solution solution = new Solution()",
				"assertArrayEquals",
			},
		},
		{
			name:     "cpp",
			language: generator.LangCpp,
			stubFile: "two_sum.cpp",
			testFile: "two_sum_test.cpp",
			stubContains: []string{
				"vector<int> twoSum",
			},
			testContains: []string{
				`#include "two_sum.cpp"`,
				"assert(",
			},
		},
		{
			name:     "javascript",
			language: generator.LangJavaScript,
			stubFile: "two_sum.js",
			testFile: "two_sum.test.js",
			stubContains: []string{
				"function twoSum(nums, target)",
				"module.exports = twoSum",
			},
			testContains: []string{
				"const twoSum = require('./two_sum')",
			},
		},
		{
			name:     "rust",
			language: generator.LangRust,
			stubFile: "two_sum.rs",
			testFile: "two_sum_test.rs",
			stubContains: []string{
				"pub struct Solution",
				"pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32>",
			},
			testContains: []string{
				`#[path = "two_sum.rs"]`,
				"use two_sum::Solution",
				"assert_eq!",
			},
		},
		{
			name:     "csharp",
			language: generator.LangCSharp,
			stubFile: "two_sum.cs",
			testFile: "two_sumTests.cs",
			stubContains: []string{
				"public class Solution",
				"public int[] TwoSum",
			},
			testContains: []string{
				"Solution solution = new Solution()",
				"Assert.Equal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			mgr := New(root, generator.New())
			p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum", Category: "arrays-hashing"}

			stubPath, testPath, err := mgr.Generate(p, tt.language)
			require.NoError(t, err)

			assert.Equal(t, filepath.Join(root, "arrays-hashing", "1-two-sum", tt.stubFile), stubPath)
			assert.Equal(t, filepath.Join(root, "arrays-hashing", "1-two-sum", tt.testFile), testPath)

			stub := readFileForTest(t, stubPath)
			for _, expected := range tt.stubContains {
				assert.Contains(t, stub, expected)
			}

			testFile := readFileForTest(t, testPath)
			for _, expected := range tt.testContains {
				assert.Contains(t, testFile, expected)
			}
			assert.NotContains(t, testFile, "TODO: add test cases")
		})
	}
}

func TestManager_Generate_GoTestsAreRunnableRedTests(t *testing.T) {
	root := t.TempDir()
	mgr := New(root, generator.New())
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum", Category: "arrays-hashing"}

	_, _, err := mgr.Generate(p, generator.LangGo)
	require.NoError(t, err)

	cmd := exec.Command("go", "test", ".")
	cmd.Dir = mgr.ProblemDir(p)
	output, err := cmd.CombinedOutput()

	require.Error(t, err, "generated stub should compile but fail the concrete example tests until implemented")
	assert.Contains(t, string(output), "got [], want [0 1]")
	assert.NotContains(t, string(output), "build failed")
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

func readFileForTest(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func allCatalogedProblems(t *testing.T) []*roadmap.Problem {
	t.Helper()
	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	seen := map[int]bool{}
	var problems []*roadmap.Problem
	for _, p := range rm.Graph.Problems {
		if !seen[p.ID] {
			seen[p.ID] = true
			problems = append(problems, p)
		}
	}

	rm2, err := catalog.LoadRoadmap("hard-mode")
	require.NoError(t, err)
	for _, p := range rm2.Graph.Problems {
		if !seen[p.ID] {
			seen[p.ID] = true
			problems = append(problems, p)
		}
	}

	sort.Slice(problems, func(i, j int) bool { return problems[i].ID < problems[j].ID })
	return problems
}

func TestManager_GenerateAllCatalogedProblems_Go(t *testing.T) {
	problems := allCatalogedProblems(t)
	require.NotEmpty(t, problems)
	t.Logf("testing generation for %d cataloged problems", len(problems))

	root := t.TempDir()
	gen := generator.New()
	mgr := New(root, gen)

	for _, p := range problems {
		t.Run(fmt.Sprintf("#%d_%s", p.ID, p.Slug), func(t *testing.T) {
			canGenerate, _, _, reason := generator.AutomationSupport(p)
			if !canGenerate {
				_, _, err := mgr.Generate(p, generator.LangGo)
				require.Error(t, err)
				assert.Contains(t, err.Error(), reason)
				return
			}

			stubPath, testPath, err := mgr.Generate(p, generator.LangGo)
			require.NoError(t, err, "Generate should not error for problem %d", p.ID)

			assert.FileExists(t, stubPath, "stub must exist for problem %d", p.ID)
			assert.FileExists(t, testPath, "test must exist for problem %d", p.ID)

			stub := readFileForTest(t, stubPath)
			testData := readFileForTest(t, testPath)

			assert.NotEmpty(t, stub, "stub must not be empty for problem %d", p.ID)
			assert.NotEmpty(t, testData, "test must not be empty for problem %d", p.ID)

			assert.Contains(t, stub, "package solution", "stub must have package declaration for problem %d", p.ID)

			assert.Contains(t, stub, "func ", "stub must contain a function for problem %d", p.ID)
			assert.Contains(t, testData, "func Test", "test must contain a test function for problem %d", p.ID)
			assert.NotContains(t, testData, "TODO: add test cases", "test must not contain placeholder for problem %d", p.ID)
		})
	}
}

func TestManager_GenerateAllCatalogedProblems_GoTestsCompile(t *testing.T) {
	problems := allCatalogedProblems(t)
	root := t.TempDir()
	gen := generator.New()
	mgr := New(root, gen)

	// Sample representative problems from each category shape
	samples := []struct {
		id   int
		name string
	}{
		{1, "two-sum (int slice return)"},
		{217, "contains-duplicate (bool return)"},
		{242, "valid-anagram (string params, bool)"},
		{121, "max-profit (int return)"},
		{20, "valid-parentheses (string param, bool)"},
		{70, "climbing-stairs (int param, int return)"},
		{206, "reverse-linked-list (ListNode)"},
		{94, "binary-tree-inorder (TreeNode)"},
		{56, "merge-intervals (2d int slice)"},
		{128, "longest-consecutive (int param, int return)"},
		{49, "group-anagrams (string slice return)"},
		{200, "number-of-islands (2d byte slice)"},
		{53, "maximum-subarray (int slice)"},
		{78, "subsets (int slice return 2d)"},
		{39, "combination-sum (int slice + int param)"},
		{48, "rotate-image (void return, in-place matrix)"},
		{73, "set-matrix-zeroes (void return, in-place matrix)"},
		{143, "reorder-list (void return, in-place ListNode)"},
		{286, "walls-and-gates (void return, in-place grid)"},
		{66, "plus-one (int slice)"},
		{202, "happy-number (int param, bool)"},
		{23, "merge-k-sorted-lists (CmpSkip)"},
		{105, "construct-binary-tree (CmpSkip)"},
		{108, "convert-sorted-array-to-bst (CmpSkip)"},
	}

	for _, s := range samples {
		t.Run(s.name, func(t *testing.T) {
			p, ok := findProblemByID(problems, s.id)
			require.True(t, ok, "problem %d must exist in catalog", s.id)

			_, _, err := mgr.Generate(p, generator.LangGo)
			require.NoError(t, err)

			cmd := exec.Command("go", "test", ".")
			cmd.Dir = mgr.ProblemDir(p)
			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			assert.NotContains(t, outputStr, "build failed", "problem %d must compile without build errors", p.ID)
		})
	}
}

func TestManager_Generate_RejectsDesignProblems(t *testing.T) {
	root := t.TempDir()
	mgr := New(root, generator.New())
	p := &roadmap.Problem{ID: 146, Title: "LRU Cache", Slug: "lru-cache", Category: "linked-list"}

	_, _, err := mgr.Generate(p, generator.LangGo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "design Problems are not supported")
}

func findProblemByID(problems []*roadmap.Problem, id int) (*roadmap.Problem, bool) {
	for _, p := range problems {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}
