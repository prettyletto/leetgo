package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	leetcodeBaseURL = "https://leetcode.com"
	submitEndpoint  = "/problems/%s/submit/"
	checkEndpoint   = "/submissions/detail/%s/check/"
)

type submitRequest struct {
	QuestionID int    `json:"question_id"`
	Lang       string `json:"lang"`
	TypedCode  string `json:"typed_code"`
}

type submitResponse struct {
	SubmissionID int `json:"submission_id"`
}

type checkResponse struct {
	State            string        `json:"state"`
	StatusCode       int           `json:"status_code"`
	StatusMsg        string        `json:"status_msg"`
	Runtime          runtimeMetric `json:"runtime"`
	StatusRuntime    runtimeMetric `json:"status_runtime"`
	Memory           memoryMetric  `json:"memory"`
	StatusMemory     memoryMetric  `json:"status_memory"`
	TotalCorrect     int           `json:"total_correct"`
	TotalTestcases   int           `json:"total_testcases"`
	CompileError     string        `json:"compile_error"`
	FullCompileError string        `json:"full_compile_error"`
	RuntimeError     string        `json:"runtime_error"`
	LastTestcase     string        `json:"last_testcase"`
}

type runtimeMetric string

func (m *runtimeMetric) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*m = runtimeMetric(s)
		return nil
	}

	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*m = runtimeMetric(fmt.Sprintf("%g ms", n))
		return nil
	}

	if string(data) == "null" {
		*m = ""
		return nil
	}

	return fmt.Errorf("unsupported metric value: %s", string(data))
}

type memoryMetric string

func (m *memoryMetric) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*m = memoryMetric(s)
		return nil
	}

	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*m = memoryMetric(formatBytesAsMB(n))
		return nil
	}

	if string(data) == "null" {
		*m = ""
		return nil
	}

	return fmt.Errorf("unsupported memory value: %s", string(data))
}

func formatBytesAsMB(bytes float64) string {
	if bytes <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f MB", bytes/1024/1024)
}

func (c *Client) submitSolution(ctx context.Context, problemID int, problemSlug string, lang string, code string) (*SubmissionResult, error) {
	if err := c.ValidateSession(ctx); err != nil {
		return nil, err
	}
	problemURL := fmt.Sprintf(leetcodeBaseURL+"/problems/%s/", problemSlug)

	reqBody := submitRequest{
		QuestionID: problemID,
		Lang:       lang,
		TypedCode:  adaptSubmissionCode(lang, problemSlug, code),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf(leetcodeBaseURL+submitEndpoint, problemSlug)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req, problemURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		c.invalidateSession()
		return nil, ErrSessionExpired
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("submit forbidden (403): %s", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("submit failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var submitResp submitResponse
	if err := json.Unmarshal(respBody, &submitResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return c.pollResult(ctx, submitResp.SubmissionID)
}

func (c *Client) pollResult(ctx context.Context, submissionID int) (*SubmissionResult, error) {
	url := fmt.Sprintf(leetcodeBaseURL+checkEndpoint, fmt.Sprintf("%d", submissionID))

	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create check request: %w", err)
		}
		c.setHeaders(req, leetcodeBaseURL)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("check result: %w", err)
		}
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			c.invalidateSession()
			return nil, ErrSessionExpired
		}
		if resp.StatusCode == http.StatusForbidden {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("submission check forbidden (403): %s", string(respBody))
		}

		var checkResp checkResponse
		if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode check response: %w", err)
		}
		resp.Body.Close()

		if checkResp.State == "SUCCESS" {
			return &SubmissionResult{
				Status:      checkResp.StatusMsg,
				StatusCode:  checkResp.StatusCode,
				Runtime:     checkResp.runtime(),
				Memory:      checkResp.memory(),
				TotalTests:  checkResp.TotalTestcases,
				PassedTests: checkResp.TotalCorrect,
				Error:       checkResp.errorDetail(),
			}, nil
		}
	}

	return nil, fmt.Errorf("submission check timed out")
}

func (r checkResponse) runtime() string {
	if r.StatusRuntime != "" {
		return string(r.StatusRuntime)
	}
	return string(r.Runtime)
}

func (r checkResponse) memory() string {
	if r.StatusMemory != "" {
		return string(r.StatusMemory)
	}
	return string(r.Memory)
}

func (r checkResponse) errorDetail() string {
	for _, detail := range []string{r.FullCompileError, r.CompileError, r.RuntimeError} {
		if detail != "" {
			return detail
		}
	}
	if r.LastTestcase != "" {
		return "Last testcase: " + r.LastTestcase
	}
	return ""
}

func (c *Client) setHeaders(req *http.Request, referer string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", leetcodeBaseURL)
	req.Header.Set("X-CSRFToken", c.session.CSRFToken)
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.session.CSRFToken})
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.session.SessionID})
}
