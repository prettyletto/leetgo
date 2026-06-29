package leetcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTMLToPlainText(t *testing.T) {
	input := `<p>Given a string <code>s</code>.</p><p><strong>Example 1:</strong></p><pre>Input: s = "ABAB", k = 2
Output: 4</pre><p>&nbsp;</p><p>Return the answer.</p>`

	got := htmlToPlainText(input)

	assert.Contains(t, got, `Given a string s.`)
	assert.Contains(t, got, `Example 1:`)
	assert.Contains(t, got, `Input: s = "ABAB", k = 2`)
	assert.Contains(t, got, `Output: 4`)
	assert.Contains(t, got, `Return the answer.`)
	assert.NotContains(t, got, `<p>`)
}
