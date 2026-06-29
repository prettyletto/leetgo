package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prettyletto/leetgo/internal/config"
)

type Session struct {
	CSRFToken       string     `json:"csrf_token"`
	SessionID       string     `json:"session_id"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	LastValidatedAt time.Time  `json:"last_validated_at,omitempty"`
	Source          string     `json:"source,omitempty"`
}

type Client struct {
	httpClient *http.Client
	session    *Session
	dataDir    string
}

var ErrSessionExpired = errors.New("Session expired. Run `leetgo auth` to reconnect")

func NewClient() (*Client, error) {
	dataDir, err := config.DataDir()
	if err != nil {
		return nil, fmt.Errorf("get data dir: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		dataDir:    dataDir,
	}

	if err := c.loadSession(); err != nil {
		return c, nil
	}

	return c, nil
}

func (c *Client) IsAuthenticated() bool {
	if c.session == nil || c.session.CSRFToken == "" || c.session.SessionID == "" {
		return false
	}
	return c.session.ExpiresAt == nil || c.session.ExpiresAt.After(time.Now())
}

func (c *Client) Authenticate(ctx context.Context) error {
	return c.browserSessionAuth(ctx)
}

func (c *Client) Submit(ctx context.Context, problemID int, problemSlug string, lang string, code string) (*SubmissionResult, error) {
	if !c.IsAuthenticated() {
		return nil, ErrSessionExpired
	}

	return c.submitSolution(ctx, problemID, problemSlug, lang, code)
}

func (c *Client) ValidateSession(ctx context.Context) error {
	if !c.IsAuthenticated() {
		return ErrSessionExpired
	}

	body, err := json.Marshal(map[string]string{
		"query": "query leetgoUserStatus { userStatus { isSignedIn username } }",
	})
	if err != nil {
		return fmt.Errorf("marshal validation request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://leetcode.com/graphql", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create validation request: %w", err)
	}
	c.setHeaders(req, "https://leetcode.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("validate Session: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.invalidateSession()
		return ErrSessionExpired
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("validate Session failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var payload struct {
		Data struct {
			UserStatus struct {
				IsSignedIn bool   `json:"isSignedIn"`
				Username   string `json:"username"`
			} `json:"userStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return fmt.Errorf("decode validation response: %w", err)
	}
	if !payload.Data.UserStatus.IsSignedIn {
		c.invalidateSession()
		return ErrSessionExpired
	}

	c.session.LastValidatedAt = time.Now()
	if err := c.saveSession(); err != nil {
		return fmt.Errorf("save validated Session: %w", err)
	}
	return nil
}

func (c *Client) sessionPath() string {
	return filepath.Join(c.dataDir, "session.json")
}

func (c *Client) loadSession() error {
	data, err := os.ReadFile(c.sessionPath())
	if err != nil {
		return err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	c.session = &session
	return nil
}

func (c *Client) saveSession() error {
	if c.session == nil {
		return os.Remove(c.sessionPath())
	}
	data, err := json.MarshalIndent(c.session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.sessionPath(), data, 0o600)
}

func (c *Client) invalidateSession() {
	c.session = nil
	_ = c.saveSession()
}

type SubmissionResult struct {
	Status         string `json:"status"`
	StatusCode     int    `json:"status_code"`
	Runtime        string `json:"runtime"`
	Memory         string `json:"memory"`
	TotalTests     int    `json:"total_tests"`
	PassedTests    int    `json:"passed_tests"`
	Error          string `json:"error,omitempty"`
	LastTestcase   string `json:"last_testcase,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
}
