package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type TypeScriptTemplate struct{}

const tsStubTmpl = `export function {{.FuncName}}({{.Params}}): {{.Return}} {
  // TODO: implement
}
`

const tsTestTmpl = `import { describe, it, expect } from 'vitest';
import { {{.FuncName}} } from './solution';

describe('{{.FuncName}}', () => {
  it('should pass example cases', () => {
    // TODO: add test cases from LeetCode examples
  });
});
`

func (t *TypeScriptTemplate) StubExt() string { return ".ts" }
func (t *TypeScriptTemplate) TestExt() string { return ".test.ts" }

func (t *TypeScriptTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", tsStubTmpl, tsData(p))
}

func (t *TypeScriptTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", tsTestTmpl, tsData(p))
}

func tsData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toCamelCase(p.Slug),
		"Params":   "nums: number[], target: number",
		"Return":   "number[]",
	}
}

func toCamelCase(slug string) string {
	result := make([]byte, 0, len(slug))
	capitalize := false
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
