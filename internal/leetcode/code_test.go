package leetcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdaptGoSubmissionCode(t *testing.T) {
	code := "package solution\n\nfunc TwoSum(nums []int, target int) []int {\n\treturn nil\n}\n"

	adapted := adaptSubmissionCode("golang", "two-sum", code)

	assert.NotContains(t, adapted, "package solution")
	assert.Contains(t, adapted, "func twoSum(nums []int, target int) []int")
	assert.NotContains(t, adapted, "func TwoSum")
}

func TestAdaptGoSubmissionCode_UsesCuratedLeetCodeFunctionName(t *testing.T) {
	code := "package solution\n\nfunc ValidAnagram(s string, t string) bool {\n\treturn true\n}\n"

	adapted := adaptSubmissionCode("golang", "valid-anagram", code)

	assert.NotContains(t, adapted, "package solution")
	assert.Contains(t, adapted, "func isAnagram(s string, t string) bool")
	assert.NotContains(t, adapted, "func ValidAnagram")
}

func TestAdaptGoSubmissionCode_RewritesOldPascalNameToCanonicalName(t *testing.T) {
	code := "package solution\n\nfunc BestTimeToBuyAndSellStock(prices []int) int {\n\treturn 0\n}\n"

	adapted := adaptSubmissionCode("golang", "best-time-to-buy-and-sell-stock", code)

	assert.Contains(t, adapted, "func maxProfit(prices []int) int")
	assert.NotContains(t, adapted, "func BestTimeToBuyAndSellStock")
}

func TestAdaptGoSubmissionCode_CuratedNameAlreadyCorrect(t *testing.T) {
	code := "package solution\n\nfunc isAnagram(s string, t string) bool {\n\treturn true\n}\n"

	adapted := adaptSubmissionCode("golang", "valid-anagram", code)

	assert.Contains(t, adapted, "func isAnagram(s string, t string) bool")
}

func TestAdaptSubmissionCode_NonGoUnchanged(t *testing.T) {
	code := "def two_sum(nums, target):\n    return []\n"

	assert.Equal(t, code, adaptSubmissionCode("java", "two-sum", code))
}

func TestAdaptPythonSubmissionCode_WrapsGeneratedFunctionInSolutionClass(t *testing.T) {
	code := "def twoSum(nums: list[int], target: int) -> list[int]:\n    return []\n"

	adapted := adaptSubmissionCode("python3", "two-sum", code)

	assert.Contains(t, adapted, "class Solution:\n    def twoSum(self, nums: list[int], target: int) -> list[int]:")
	assert.Contains(t, adapted, "        return []")
}

func TestAdaptTypeScriptSubmissionCode_RemovesExport(t *testing.T) {
	code := "export function twoSum(nums: number[], target: number): number[] {\n  return [];\n}\n"

	adapted := adaptSubmissionCode("typescript", "two-sum", code)

	assert.Contains(t, adapted, "function twoSum(nums: number[], target: number): number[]")
	assert.NotContains(t, adapted, "export")
}

func TestAdaptJavaScriptSubmissionCode_RemovesModuleExport(t *testing.T) {
	code := "function twoSum(nums, target) {\n  return [];\n}\n\nmodule.exports = twoSum;\n"

	adapted := adaptSubmissionCode("javascript", "two-sum", code)

	assert.Contains(t, adapted, "function twoSum(nums, target)")
	assert.NotContains(t, adapted, "module.exports")
}
