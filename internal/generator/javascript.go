package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type JavaScriptTemplate struct{}

const jsStubTmpl = `/**
 * @param {number[]} nums
 * @param {number} target
 * @return {number[]}
 */
function {{.FuncName}}(nums, target) {
  // TODO: implement
}

module.exports = {{.FuncName}};
`

const jsTestTmpl = `const {{.FuncName}} = require('./solution');

describe('{{.FuncName}}', () => {
  test('passes example cases', () => {
    // TODO: add test cases from LeetCode examples
  });
});
`

func (j *JavaScriptTemplate) StubExt() string { return ".js" }
func (j *JavaScriptTemplate) TestExt() string { return ".test.js" }

func (j *JavaScriptTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", jsStubTmpl, jsData(p))
}

func (j *JavaScriptTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", jsTestTmpl, jsData(p))
}

func jsData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toCamelCase(p.Slug),
	}
}
