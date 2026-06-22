package leetcode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_IsAuthenticated_NoSession(t *testing.T) {
	c := &Client{}
	assert.False(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_ExpiredSession(t *testing.T) {
	c := &Client{
		session: &Session{
			CSRFToken: "test",
			SessionID: "test",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		},
	}
	assert.False(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_ValidSession(t *testing.T) {
	c := &Client{
		session: &Session{
			CSRFToken: "test",
			SessionID: "test",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		},
	}
	assert.True(t, c.IsAuthenticated())
}

func TestSubmissionResult(t *testing.T) {
	result := &SubmissionResult{
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "4 ms",
		Memory:      "3.2 MB",
		TotalTests:  50,
		PassedTests: 50,
	}

	assert.Equal(t, "Accepted", result.Status)
	assert.Equal(t, 50, result.PassedTests)
}
