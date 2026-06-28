package leetcode

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitRequestUsesQuestionID(t *testing.T) {
	body, err := json.Marshal(submitRequest{
		QuestionID: 1,
		Lang:       "golang",
		TypedCode:  "func twoSum() {}",
	})
	require.NoError(t, err)

	assert.JSONEq(t, `{"question_id":1,"lang":"golang","typed_code":"func twoSum() {}"}`, string(body))
	assert.NotContains(t, string(body), "question_slug")
}

func TestCheckResponseAllowsNumericMetrics(t *testing.T) {
	var resp checkResponse
	err := json.Unmarshal([]byte(`{
		"state":"SUCCESS",
		"status_code":10,
		"status_msg":"Accepted",
		"runtime":0,
		"memory":18200000,
		"total_correct":63,
		"total_testcases":63
	}`), &resp)

	require.NoError(t, err)
	assert.Equal(t, "0 ms", string(resp.Runtime))
	assert.Equal(t, "17.36 MB", string(resp.Memory))
}

func TestCheckResponseUsesStatusRuntime(t *testing.T) {
	var resp checkResponse
	err := json.Unmarshal([]byte(`{
		"state":"SUCCESS",
		"status_code":10,
		"status_msg":"Accepted",
		"status_runtime":"3 ms",
		"memory":5412000,
		"total_correct":63,
		"total_testcases":63
	}`), &resp)

	require.NoError(t, err)
	assert.Equal(t, "3 ms", resp.runtime())
	assert.Equal(t, "5.16 MB", resp.memory())
}

func TestCheckResponseErrorDetail(t *testing.T) {
	resp := checkResponse{
		CompileError:     "short compile error",
		FullCompileError: "full compile error",
		RuntimeError:     "runtime error",
	}

	assert.Equal(t, "full compile error", resp.errorDetail())
}

func TestCheckResponseKeepsFailedCaseFields(t *testing.T) {
	var resp checkResponse
	err := json.Unmarshal([]byte(`{
		"state":"SUCCESS",
		"status_code":11,
		"status_msg":"Wrong Answer",
		"last_testcase":"[3,2,4]\n6",
		"expected_output":"[1,2]"
	}`), &resp)

	require.NoError(t, err)
	assert.Equal(t, "[3,2,4]\n6", resp.LastTestcase)
	assert.Equal(t, "[1,2]", resp.ExpectedOutput)
}
