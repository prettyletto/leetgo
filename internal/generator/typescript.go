package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type TypeScriptTemplate struct{}

func (t *TypeScriptTemplate) StubExt() string { return ".ts" }
func (t *TypeScriptTemplate) TestExt() string { return ".test.ts" }

func (t *TypeScriptTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderTSStub(spec), nil
}

func (t *TypeScriptTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderTSTest(spec), nil
}

func tsType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "number"
	case KindIntSlice:
		return "number[]"
	case KindIntSliceSlice:
		return "number[][]"
	case KindBool:
		return "boolean"
	case KindString:
		return "string"
	case KindStringSlice:
		return "string[]"
	case KindStringSliceSlice:
		return "string[][]"
	case KindByteSliceSlice:
		return "string[][]"
	case KindFloat64:
		return "number"
	case KindListNode:
		return "ListNode | null"
	case KindTreeNode:
		return "TreeNode | null"
	default:
		return "any"
	}
}

func tsZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindBool:
		return "false"
	case KindString:
		return `""`
	case KindIntSlice:
		return "[]"
	case KindIntSliceSlice:
		return "[]"
	case KindStringSlice:
		return "[]"
	default:
		return "null"
	}
}

func renderTSStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.TSFuncName()
	params := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		params[i] = fmt.Sprintf("%s: %s", p.Name, tsType(p.Type))
	}
	buf.WriteString(fmt.Sprintf("export function %s(%s): %s {\n", funcName, strings.Join(params, ", "), tsType(spec.Return.Type)))
	buf.WriteString(fmt.Sprintf("  // TODO: implement\n"))
	buf.WriteString(fmt.Sprintf("  return %s;\n", tsZero(spec.Return.Type)))
	buf.WriteString("}\n")
	return buf.Bytes()
}

func renderTSTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.TSFuncName()
	fileBase := spec.FileBase()
	buf.WriteString("import { describe, it, expect } from 'vitest';\n")
	if spec.NeedsListNode || spec.NeedsTreeNode {
		buf.WriteString(fmt.Sprintf("import { %s", funcName))
		if spec.NeedsListNode {
			buf.WriteString(", ListNode")
		}
		if spec.NeedsTreeNode {
			buf.WriteString(", TreeNode")
		}
		buf.WriteString(fmt.Sprintf(" } from './%s';\n", fileBase))
	} else {
		buf.WriteString(fmt.Sprintf("import { %s } from './%s';\n", funcName, fileBase))
	}
	buf.WriteString(fmt.Sprintf("\ndescribe('%s', () => {\n", funcName))
	buf.WriteString("  it('should pass example cases', () => {\n")
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
