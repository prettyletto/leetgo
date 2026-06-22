package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type PythonTemplate struct{}

const pythonStubTmpl = `def {{.FuncName}}({{.Params}}) -> {{.Return}}:
    # TODO: implement
    pass
`

const pythonTestTmpl = `import pytest
from solution import {{.FuncName}}

@pytest.mark.parametrize("input,expected", [
    # TODO: add test cases from LeetCode examples
])
def test_{{.FuncName}}(input, expected):
    assert {{.FuncName}}(*input) == expected
`

func (p *PythonTemplate) StubExt() string { return ".py" }
func (p *PythonTemplate) TestExt() string { return "_test.py" }

func (p *PythonTemplate) RenderStub(prob *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", pythonStubTmpl, pythonData(prob))
}

func (p *PythonTemplate) RenderTest(prob *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", pythonTestTmpl, pythonData(prob))
}

func pythonData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toSnakeCase(p.Slug),
		"Params":   "nums: list[int], target: int",
		"Return":   "list[int]",
	}
}

func toSnakeCase(slug string) string {
	result := make([]byte, 0, len(slug))
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' {
			result = append(result, '_')
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}
