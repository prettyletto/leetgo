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
	expiresAt := time.Now().Add(-1 * time.Hour)
	c := &Client{
		session: &Session{
			CSRFToken: "test",
			SessionID: "test",
			ExpiresAt: &expiresAt,
		},
	}
	assert.False(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_ValidSession(t *testing.T) {
	expiresAt := time.Now().Add(1 * time.Hour)
	c := &Client{
		session: &Session{
			CSRFToken: "test",
			SessionID: "test",
			ExpiresAt: &expiresAt,
		},
	}
	assert.True(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_UnknownExpiryWithTokens(t *testing.T) {
	c := &Client{
		session: &Session{
			CSRFToken: "test",
			SessionID: "test",
		},
	}
	assert.True(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_MissingCSRF(t *testing.T) {
	c := &Client{
		session: &Session{
			SessionID: "test",
		},
	}
	assert.False(t, c.IsAuthenticated())
}

func TestClient_IsAuthenticated_MissingSessionID(t *testing.T) {
	c := &Client{
		session: &Session{
			CSRFToken: "test",
		},
	}
	assert.False(t, c.IsAuthenticated())
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
