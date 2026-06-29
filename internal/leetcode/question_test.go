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

func TestQuestionMetadataUsesLeetCodeInternalQuestionID(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	c := &Client{
		dataDir: t.TempDir(),
		session: &Session{
			CSRFToken: "csrf",
			SessionID: "session",
			ExpiresAt: &expiresAt,
		},
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, graphqlEndpoint, req.URL.String())
			assert.Equal(t, "https://leetcode.com/problems/binary-search/", req.Header.Get("Referer"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"data": {"question": {"questionId": "792", "questionFrontendId": "704", "titleSlug": "binary-search"}}
				}`)),
				Header: make(http.Header),
			}, nil
		})},
	}

	metadata, err := c.questionMetadata(context.Background(), "binary-search")

	require.NoError(t, err)
	assert.Equal(t, 792, metadata.QuestionID)
	assert.Equal(t, 704, metadata.QuestionFrontendID)
	assert.Equal(t, "binary-search", metadata.TitleSlug)
}
