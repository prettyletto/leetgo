package generator

import (
	"bytes"
	"text/template"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type GoTemplate struct{}

const goStubTmpl = `package solution

func {{.FuncName}}({{.Params}}) {{.Return}} {
	// TODO: implement
}
`

const goTestTmpl = `package solution

import "testing"

func Test{{.FuncName}}(t *testing.T) {
	tests := []struct {
		name   string
		input  {{.InputType}}
		expect {{.ReturnType}}
	}{
		// TODO: add test cases from LeetCode examples
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := {{.FuncName}}(tt.input)
			if got != tt.expect {
				t.Errorf("got %v, want %v", got, tt.expect)
			}
		})
	}
}
`

func (g *GoTemplate) StubExt() string { return ".go" }
func (g *GoTemplate) TestExt() string { return "_test.go" }

func (g *GoTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", goStubTmpl, goData(p))
}

func (g *GoTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", goTestTmpl, goData(p))
}

func goData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName":   toPascalCase(p.Slug),
		"Params":     "nums []int, target int",
		"Return":     "[]int",
		"InputType":  "struct{ nums []int; target int }",
		"ReturnType": "[]int",
	}
}

func renderTemplate(name, tmpl string, data map[string]string) ([]byte, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toPascalCase(slug string) string {
	result := make([]byte, 0, len(slug))
	capitalize := true
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' || c == '_' {
			capitalize = true
			continue
		}
		if capitalize && c >= 'a' && c <= 'z' {
			c -= 32
		}
		result = append(result, c)
		capitalize = false
	}
	return string(result)
}
