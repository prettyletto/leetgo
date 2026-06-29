package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const graphqlEndpoint = leetcodeBaseURL + "/graphql"

type ProblemDescription struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	ContentHTML string `json:"content_html"`
	ContentText string `json:"content_text"`
}

func (c *Client) ProblemDescription(ctx context.Context, slug string) (*ProblemDescription, error) {
	if cached, err := c.cachedProblemDescription(slug); err == nil && cached.ContentText != "" {
		return cached, nil
	}

	desc, err := c.fetchProblemDescription(ctx, slug)
	if err != nil {
		return nil, err
	}
	_ = c.cacheProblemDescription(desc)
	return desc, nil
}

func (c *Client) fetchProblemDescription(ctx context.Context, slug string) (*ProblemDescription, error) {
	body, err := json.Marshal(map[string]any{
		"query": `query questionData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    title
    titleSlug
    content
  }
}`,
		"variables": map[string]string{"titleSlug": slug},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch LeetCode description failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var payload struct {
		Data struct {
			Question struct {
				Title     string `json:"title"`
				TitleSlug string `json:"titleSlug"`
				Content   string `json:"content"`
			} `json:"question"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, err
	}
	if len(payload.Errors) > 0 {
		return nil, fmt.Errorf("fetch LeetCode description: %s", payload.Errors[0].Message)
	}
	if payload.Data.Question.Content == "" {
		return nil, fmt.Errorf("LeetCode description is empty for %s", slug)
	}

	return &ProblemDescription{
		Slug:        payload.Data.Question.TitleSlug,
		Title:       payload.Data.Question.Title,
		ContentHTML: payload.Data.Question.Content,
		ContentText: htmlToPlainText(payload.Data.Question.Content),
	}, nil
}

func (c *Client) cachedProblemDescription(slug string) (*ProblemDescription, error) {
	data, err := os.ReadFile(c.problemDescriptionPath(slug))
	if err != nil {
		return nil, err
	}
	var desc ProblemDescription
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, err
	}
	return &desc, nil
}

func (c *Client) cacheProblemDescription(desc *ProblemDescription) error {
	if desc == nil || desc.Slug == "" {
		return nil
	}
	dir := filepath.Dir(c.problemDescriptionPath(desc.Slug))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.problemDescriptionPath(desc.Slug), data, 0o644)
}

func (c *Client) problemDescriptionPath(slug string) string {
	return filepath.Join(c.dataDir, "descriptions", slug+".json")
}

var htmlBlockBreaks = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|pre|blockquote|h[1-6])[^>]*>`)
var htmlTags = regexp.MustCompile(`(?s)<[^>]+>`)
var blankLines = regexp.MustCompile(`\n{3,}`)

func htmlToPlainText(input string) string {
	text := htmlBlockBreaks.ReplaceAllString(input, "\n")
	text = htmlTags.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimFunc(line, unicode.IsSpace)
		if strings.HasPrefix(line, " ") {
			line = strings.TrimSpace(line)
		}
		lines = append(lines, line)
	}
	text = strings.Join(lines, "\n")
	text = blankLines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
