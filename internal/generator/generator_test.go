package generator

import (
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &GoTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "func TwoSum")
	assert.Contains(t, string(stub), "nums []int")
}

func TestGoTemplate_RenderTest(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &GoTemplate{}

	test, err := tmpl.RenderTest(p)
	require.NoError(t, err)
	assert.Contains(t, string(test), "TestTwoSum")
}

func TestPythonTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &PythonTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "def two_sum")
}

func TestTypeScriptTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &TypeScriptTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "function twoSum")
}

func TestJavaTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &JavaTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "public int[] twoSum")
}

func TestCppTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &CppTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "vector<int> twoSum")
}

func TestJavaScriptTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &JavaScriptTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "function twoSum")
}

func TestRustTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &RustTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "pub fn two_sum")
}

func TestCSharpTemplate_RenderStub(t *testing.T) {
	p := &roadmap.Problem{ID: 1, Title: "Two Sum", Slug: "two-sum"}
	tmpl := &CSharpTemplate{}

	stub, err := tmpl.RenderStub(p)
	require.NoError(t, err)
	assert.Contains(t, string(stub), "public int[] TwoSum")
}

func TestGenerator_Languages(t *testing.T) {
	gen := New()
	languages := gen.Languages()
	assert.Contains(t, languages, LangCpp)
	assert.Contains(t, languages, LangJavaScript)
	assert.Contains(t, languages, LangRust)
	assert.Contains(t, languages, LangCSharp)
}

func TestToPascalCase(t *testing.T) {
	assert.Equal(t, "TwoSum", toPascalCase("two-sum"))
	assert.Equal(t, "LongestSubstring", toPascalCase("longest-substring"))
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "two_sum", toSnakeCase("two-sum"))
}

func TestToCamelCase(t *testing.T) {
	assert.Equal(t, "twoSum", toCamelCase("two-sum"))
}
