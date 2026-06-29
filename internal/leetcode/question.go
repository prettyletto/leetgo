package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type questionMetadata struct {
	QuestionID         int
	QuestionFrontendID int
	TitleSlug          string
}

func (c *Client) questionMetadata(ctx context.Context, slug string) (*questionMetadata, error) {
	body, err := json.Marshal(map[string]any{
		"query": `query questionSubmissionMeta($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    questionFrontendId
    titleSlug
  }
}`,
		"variables": map[string]string{"titleSlug": slug},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal question metadata request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create question metadata request: %w", err)
	}
	c.setHeaders(req, leetcodeBaseURL+"/problems/"+slug+"/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch question metadata: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.invalidateSession()
		return nil, ErrSessionExpired
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch question metadata failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var payload struct {
		Data struct {
			Question *struct {
				QuestionID         string `json:"questionId"`
				QuestionFrontendID string `json:"questionFrontendId"`
				TitleSlug          string `json:"titleSlug"`
			} `json:"question"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("decode question metadata response: %w", err)
	}
	if len(payload.Errors) > 0 {
		return nil, fmt.Errorf("fetch question metadata: %s", payload.Errors[0].Message)
	}
	if payload.Data.Question == nil {
		return nil, fmt.Errorf("question metadata not found for %s", slug)
	}

	questionID, err := strconv.Atoi(payload.Data.Question.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("parse questionId %q: %w", payload.Data.Question.QuestionID, err)
	}
	frontendID, _ := strconv.Atoi(payload.Data.Question.QuestionFrontendID)
	return &questionMetadata{
		QuestionID:         questionID,
		QuestionFrontendID: frontendID,
		TitleSlug:          payload.Data.Question.TitleSlug,
	}, nil
}
