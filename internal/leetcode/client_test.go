package leetcode

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

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

func TestClient_ValidateSessionConfirmsSignedInUser(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	c := &Client{
		dataDir: t.TempDir(),
		session: &Session{
			CSRFToken: "csrf",
			SessionID: "session",
			ExpiresAt: &expiresAt,
		},
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://leetcode.com/graphql", req.URL.String())
			assert.Equal(t, "csrf", req.Header.Get("X-CSRFToken"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data":{"userStatus":{"isSignedIn":true,"username":"ada"}}}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	err := c.ValidateSession(context.Background())

	require.NoError(t, err)
	assert.False(t, c.session.LastValidatedAt.IsZero())
}

func TestClient_ValidateSessionInvalidatesForbiddenSession(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	c := &Client{
		dataDir: t.TempDir(),
		session: &Session{
			CSRFToken: "csrf",
			SessionID: "session",
			ExpiresAt: &expiresAt,
		},
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader("forbidden")),
				Header:     make(http.Header),
			}, nil
		})},
	}

	err := c.ValidateSession(context.Background())

	require.ErrorIs(t, err, ErrSessionExpired)
	assert.Nil(t, c.session)
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
