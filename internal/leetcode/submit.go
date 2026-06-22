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
	Lang        string `json:"lang"`
	TypedCode   string `json:"typed_code"`
	ProblemSlug string `json:"question_slug"`
}

type submitResponse struct {
	SubmissionID int `json:"submission_id"`
}

type checkResponse struct {
	State          string `json:"state"`
	StatusCode     int    `json:"status_code"`
	StatusMsg      string `json:"status_msg"`
	Runtime        string `json:"runtime"`
	Memory         string `json:"memory"`
	TotalCorrect   int    `json:"total_correct"`
	TotalTestcases int    `json:"total_testcases"`
}

func (c *Client) submitSolution(ctx context.Context, problemSlug string, lang string, code string) (*SubmissionResult, error) {
	reqBody := submitRequest{
		Lang:        lang,
		TypedCode:   code,
		ProblemSlug: problemSlug,
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

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("submit failed (%d): %s", resp.StatusCode, string(body))
	}

	var submitResp submitResponse
	if err := json.NewDecoder(resp.Body).Decode(&submitResp); err != nil {
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
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("check result: %w", err)
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
				Runtime:     checkResp.Runtime,
				Memory:      checkResp.Memory,
				TotalTests:  checkResp.TotalTestcases,
				PassedTests: checkResp.TotalCorrect,
			}, nil
		}
	}

	return nil, fmt.Errorf("submission check timed out")
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", leetcodeBaseURL)
	req.Header.Set("X-CSRFToken", c.session.CSRFToken)
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.session.CSRFToken})
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.session.SessionID})
}
