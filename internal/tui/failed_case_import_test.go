package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendFailedLeetCodeCase_GoTwoSum(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	testPath := filepath.Join(t.TempDir(), "two_sum_test.go")
	testContent := `package solution

import (
	"reflect"
	"testing"
)

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		target int
		expect []int
	}{
		{name: "example 1", nums: []int{2, 7, 11, 15}, target: 9, expect: []int{0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TwoSum(tt.nums, tt.target)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("got %v, want %v", got, tt.expect)
			}
		})
	}
}
`
	require.NoError(t, os.WriteFile(testPath, []byte(testContent), 0o644))

	msg, err := appendFailedLeetCodeCase(problem, "go", testPath, &leetcode.SubmissionResult{
		LastTestcase:   "[3,2,4]\n6",
		ExpectedOutput: "[1,2]",
	})

	require.NoError(t, err)
	assert.Contains(t, msg, "Added")
	updated, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), `{name: "leetcode failed", nums: []int{3,2,4}, target: 6, expect: []int{1,2}},`)
}

func TestLeetCodeLiteralToGo(t *testing.T) {
	tests := []struct {
		name  string
		value string
		kind  generator.ValueKind
		want  string
	}{
		{name: "int", value: "42", kind: generator.KindInt, want: "42"},
		{name: "bool", value: "false", kind: generator.KindBool, want: "false"},
		{name: "string", value: `"abc"`, kind: generator.KindString, want: `"abc"`},
		{name: "int slice", value: "[1,2,3]", kind: generator.KindIntSlice, want: "[]int{1,2,3}"},
		{name: "nested int slice", value: "[[1,2],[3,4]]", kind: generator.KindIntSliceSlice, want: "[][]int{{1,2},{3,4}}"},
		{name: "string slice", value: `["a","b"]`, kind: generator.KindStringSlice, want: `[]string{"a","b"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := leetCodeLiteralToGo(tt.value, tt.kind)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
