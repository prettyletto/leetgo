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

func TestScrapedExamplesFromDescription(t *testing.T) {
	spec, ok := generator.SpecForProblem(&roadmap.Problem{ID: 424, Slug: "longest-repeating-character-replacement"})
	require.True(t, ok)
	description := `You are given a string s and an integer k.

Example 1:
Input: s = "ABAB", k = 2
Output: 4
Explanation: Replace the two A's with two B's or vice versa.

Example 2:
Input: s = "AABABBA", k = 1
Output: 4

Constraints:`

	examples, err := scrapedExamplesFromDescription(spec, description)
	require.NoError(t, err)
	require.Len(t, examples, 2)
	assert.Equal(t, "leetcode example 1", examples[0].Input["_name"])
	assert.Equal(t, `"ABAB"`, examples[0].Input["s"])
	assert.Equal(t, "2", examples[0].Input["k"])
	assert.Equal(t, "4", examples[0].Expect)
	assert.Equal(t, `"AABABBA"`, examples[1].Input["s"])
	assert.Equal(t, "1", examples[1].Input["k"])
	assert.Equal(t, "4", examples[1].Expect)
}

func TestAppendScrapedLeetCodeExamples(t *testing.T) {
	problem := &roadmap.Problem{ID: 424, Slug: "longest-repeating-character-replacement"}
	spec, ok := generator.SpecForProblem(problem)
	require.True(t, ok)
	examples, err := scrapedExamplesFromDescription(spec, `Example 1:
Input: s = "AABABBA", k = 1
Output: 4`)
	require.NoError(t, err)
	testPath := filepath.Join(t.TempDir(), "longest_repeating_character_replacement_test.go")
	testContent := `package solution

import "testing"

func TestCharacterReplacement(t *testing.T) {
	tests := []struct {
		name string
		s string
		k int
		expect int
	}{
		{name: "curated", s: "AAAA", k: 0, expect: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
	}
}
`
	require.NoError(t, os.WriteFile(testPath, []byte(testContent), 0o644))

	msg, err := appendScrapedLeetCodeExamples(problem, "go", testPath, examples)
	require.NoError(t, err)
	assert.Contains(t, msg, "Added 1")

	updated, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), `{name: "leetcode example 1", s: "AABABBA", k: 1, expect: 4},`)
}

func TestAppendScrapedLeetCodeExamplesSkipsCuratedDuplicate(t *testing.T) {
	problem := &roadmap.Problem{ID: 424, Slug: "longest-repeating-character-replacement"}
	spec, ok := generator.SpecForProblem(problem)
	require.True(t, ok)
	examples, err := scrapedExamplesFromDescription(spec, `Example 1:
Input: s = "ABAB", k = 2
Output: 4

Example 2:
Input: s = "AABABBA", k = 1
Output: 4`)
	require.NoError(t, err)
	testPath := filepath.Join(t.TempDir(), "longest_repeating_character_replacement_test.go")
	testContent := `package solution

import "testing"

func TestCharacterReplacement(t *testing.T) {
	tests := []struct {
		name string
		s string
		k int
		expect int
	}{
		{name: "example 1", s: "ABAB", k: 2, expect: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
	}
}
`
	require.NoError(t, os.WriteFile(testPath, []byte(testContent), 0o644))

	msg, err := appendScrapedLeetCodeExamples(problem, "go", testPath, examples)
	require.NoError(t, err)
	assert.Contains(t, msg, "Added 1")

	updated, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.NotContains(t, string(updated), `{name: "leetcode example 1", s: "ABAB", k: 2, expect: 4},`)
	assert.Contains(t, string(updated), `{name: "leetcode example 2", s: "AABABBA", k: 1, expect: 4},`)
}

func TestSplitTopLevelKeepsNestedCommas(t *testing.T) {
	parts := splitTopLevel(`nums = [1,2,3], pairs = [[1,2],[3,4]], s = "a,b"`, ',')

	assert.Equal(t, []string{`nums = [1,2,3]`, `pairs = [[1,2],[3,4]]`, `s = "a,b"`}, parts)
}
