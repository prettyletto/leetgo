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

func TestAdaptSubmissionCode_NonGoUnchanged(t *testing.T) {
	code := "def two_sum(nums, target):\n    return []\n"

	assert.Equal(t, code, adaptSubmissionCode("python3", "two-sum", code))
}
