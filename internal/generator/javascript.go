package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type JavaScriptTemplate struct{}

func (j *JavaScriptTemplate) StubExt() string { return ".js" }
func (j *JavaScriptTemplate) TestExt() string { return ".test.js" }

func (j *JavaScriptTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderJSStub(spec), nil
}

func (j *JavaScriptTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderJSTest(spec), nil
}

func jsZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindBool:
		return "false"
	case KindString:
		return `""`
	default:
		return "null"
	}
}

func renderJSStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.TSFuncName()
	paramNames := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		paramNames[i] = p.Name
	}
	buf.WriteString(fmt.Sprintf("function %s(%s) {\n", funcName, strings.Join(paramNames, ", ")))
	buf.WriteString("  // TODO: implement\n")
	buf.WriteString(fmt.Sprintf("  return %s;\n", jsZero(spec.Return.Type)))
	buf.WriteString("}\n\n")
	buf.WriteString(fmt.Sprintf("module.exports = %s;\n", funcName))
	return buf.Bytes()
}

func renderJSTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.TSFuncName()
	fileBase := spec.FileBase()
	buf.WriteString(fmt.Sprintf("const %s = require('./%s');\n\n", funcName, fileBase))
	if spec.NeedsListNode {
		buf.WriteString("function ListNode(val, next) { this.val = (val===undefined ? 0 : val); this.next = (next===undefined ? null : next); }\n\n")
	}
	buf.WriteString(fmt.Sprintf("describe('%s', () => {\n", funcName))
	buf.WriteString("  test('passes example cases', () => {\n")
	for _, ex := range spec.Examples {
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		buf.WriteString(fmt.Sprintf("    expect(%s(%s)).toEqual(%s);\n", funcName, strings.Join(args, ", "), ex.Expect))
	}
	buf.WriteString("  });\n")
	buf.WriteString("});\n")
	return buf.Bytes()
}
