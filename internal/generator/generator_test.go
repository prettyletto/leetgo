package generator

import (
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGo_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func TwoSum")
	assert.Contains(t, content, "nums []int, target int")
	assert.Contains(t, content, "[]int")
}

func TestGo_Test_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &GoTemplate{}
	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	content := string(testFile)
	assert.Contains(t, content, "func TestTwoSum")
	assert.Contains(t, content, "nums: []int{2,7,11,15}")
	assert.Contains(t, content, "expect: []int{0,1}")
	assert.NotContains(t, content, "TODO: add test cases")
}

func TestGo_Stub_ContainsDuplicate(t *testing.T) {
	p := &roadmap.Problem{ID: 217, Slug: "contains-duplicate"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func ContainsDuplicate")
	assert.Contains(t, content, "nums []int) bool")
	assert.Contains(t, content, "return false")
}

func TestGo_Test_ContainsDuplicate(t *testing.T) {
	p := &roadmap.Problem{ID: 217, Slug: "contains-duplicate"}
	tmpl := &GoTemplate{}
	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	content := string(testFile)
	assert.Contains(t, content, "func TestContainsDuplicate")
	assert.Contains(t, content, "nums: []int{1,2,3,1}")
	assert.Contains(t, content, "expect: true")
}

func TestGo_Stub_ClimbingStairs(t *testing.T) {
	p := &roadmap.Problem{ID: 70, Slug: "climbing-stairs"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func ClimbingStairs")
	assert.Contains(t, content, "n int) int")
}

func TestGo_Test_ClimbingStairs(t *testing.T) {
	p := &roadmap.Problem{ID: 70, Slug: "climbing-stairs"}
	tmpl := &GoTemplate{}
	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	content := string(testFile)
	assert.Contains(t, content, "func TestClimbingStairs")
	assert.Contains(t, content, "expect: 2")
	assert.Contains(t, content, "expect: 3")
}

func TestGo_Stub_GroupAnagrams(t *testing.T) {
	p := &roadmap.Problem{ID: 49, Slug: "group-anagrams"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func GroupAnagrams")
	assert.Contains(t, content, "strs []string) [][]string")
}

func TestGo_Stub_ReverseLinkedList(t *testing.T) {
	p := &roadmap.Problem{ID: 206, Slug: "reverse-linked-list"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func ReverseLinkedList")
	assert.Contains(t, content, "head *ListNode) *ListNode")
	assert.Contains(t, content, "type ListNode struct")
}

func TestGo_Test_ReverseLinkedList(t *testing.T) {
	p := &roadmap.Problem{ID: 206, Slug: "reverse-linked-list"}
	tmpl := &GoTemplate{}
	// ListNode is defined in stub, test references it
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "type ListNode struct")

	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	content := string(testFile)
	assert.Contains(t, content, "func TestReverseLinkedList")
	assert.Contains(t, content, "*ListNode")
}

func TestGo_Stub_BinaryTreeInorderTraversal(t *testing.T) {
	p := &roadmap.Problem{ID: 94, Slug: "binary-tree-inorder-traversal"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func BinaryTreeInorderTraversal")
	assert.Contains(t, content, "root *TreeNode) []int")
	assert.Contains(t, content, "type TreeNode struct")
}

func TestGo_DesignProblem_Skipped(t *testing.T) {
	p := &roadmap.Problem{ID: 146, Slug: "lru-cache"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "lru-cache requires class")

	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	assert.Contains(t, string(testFile), "SKIPPED: design problems require class/struct generation")
}

func TestGo_Stub_MergeIntervals(t *testing.T) {
	p := &roadmap.Problem{ID: 56, Slug: "merge-intervals"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func MergeIntervals")
	assert.Contains(t, content, "intervals [][]int) [][]int")
}

func TestGo_Test_SlidingWindow(t *testing.T) {
	p := &roadmap.Problem{ID: 121, Slug: "best-time-to-buy-and-sell-stock"}
	tmpl := &GoTemplate{}
	testFile, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	content := string(testFile)
	assert.Contains(t, content, "func TestBestTimeToBuyAndSellStock")
	assert.Contains(t, content, "prices: []int{7,1,5,3,6,4}")
	assert.Contains(t, content, "expect: 5")
}

func TestGo_Stub_WordLadder(t *testing.T) {
	p := &roadmap.Problem{ID: 127, Slug: "word-ladder"}
	tmpl := &GoTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	content := string(stub)
	assert.Contains(t, content, "func WordLadder")
	assert.Contains(t, content, "beginWord string, endWord string, wordList []string) int")
}

func TestAllCatalogedProblemsHaveSpecs(t *testing.T) {
	knownIDs := []int{1, 217, 242, 49, 347, 238, 36, 128, 271, 454,
		125, 167, 15, 11, 42,
		121, 3, 424, 567, 76, 239,
		20, 155, 150, 22, 739, 853, 84,
		704, 74, 875, 153, 33, 981, 4,
		206, 21, 141, 143, 19, 138, 2, 287, 146, 23, 25,
		94, 104, 226, 101, 108, 100, 110, 543, 105, 102, 98, 230, 199, 124, 297,
		208, 211, 212,
		703, 1046, 973, 215, 621, 355, 295,
		78, 39, 46, 90, 40, 17, 79, 131, 51,
		200, 133, 695, 417, 994, 207, 210, 684, 323, 261, 127,
		332, 743, 778, 787,
		70, 198, 213, 5, 647, 91, 322, 152, 139, 300, 416, 309, 494, 1143, 72, 518,
		53, 55, 45, 134, 846, 1899, 678,
		56, 57, 435, 252, 253, 986,
		48, 54, 73, 202, 66, 50, 43, 286,
		10, 312,
	}

	for _, id := range knownIDs {
		_, ok := problemSpecs[id]
		assert.True(t, ok, "problem %d should have a spec", id)
	}
	assert.Equal(t, len(knownIDs), len(problemSpecs), "all known problems should have specs in registry")
}

func TestPython_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &PythonTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "def two_sum")
	assert.Contains(t, string(stub), "nums: list[int], target: int")
	assert.Contains(t, string(stub), "list[int]")
}

func TestTypeScript_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &TypeScriptTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "function twoSum")
	assert.Contains(t, string(stub), "nums: number[], target: number")
	assert.Contains(t, string(stub), "number[]")
}

func TestJava_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &JavaTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "public int[] twoSum")
	assert.Contains(t, string(stub), "int[] nums, int target")
}

func TestCpp_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &CppTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "vector<int> twoSum")
}

func TestJavaScript_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &JavaScriptTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "function twoSum")
	assert.Contains(t, string(stub), "module.exports = twoSum")
}

func TestRust_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &RustTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "pub fn two_sum")
	assert.Contains(t, string(stub), "nums: Vec<i32>, target: i32")
	assert.Contains(t, string(stub), "Vec<i32>")
}

func TestCSharp_Stub_TwoSum(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Slug: "two-sum"}
	tmpl := &CSharpTemplate{}
	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "public int[] TwoSum")
}

func TestGenerator_Languages(t *testing.T) {
	gen := New()
	languages := gen.Languages()
	assert.Contains(t, languages, LangCpp)
	assert.Contains(t, languages, LangJavaScript)
	assert.Contains(t, languages, LangRust)
	assert.Contains(t, languages, LangCSharp)
}

func TestToPascalCase(t *testing.T) {
	assert.Equal(t, "TwoSum", toPascalCase("two-sum"))
	assert.Equal(t, "LongestSubstring", toPascalCase("longest-substring"))
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "two_sum", toSnakeCase("two-sum"))
}

func TestToCamelCase(t *testing.T) {
	assert.Equal(t, "twoSum", toCamelCase("two-sum"))
}
