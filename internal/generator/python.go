package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type PythonTemplate struct{}

func (p *PythonTemplate) StubExt() string { return ".py" }
func (p *PythonTemplate) TestExt() string { return "_test.py" }

func (p *PythonTemplate) RenderStub(prob *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(prob)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", prob.ID)
	}
	return renderPythonStub(spec), nil
}

func (p *PythonTemplate) RenderTest(prob *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(prob)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", prob.ID)
	}
	return renderPythonTest(spec), nil
}

func pyType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "int"
	case KindIntSlice:
		return "list[int]"
	case KindIntSliceSlice:
		return "list[list[int]]"
	case KindBool:
		return "bool"
	case KindString:
		return "str"
	case KindStringSlice:
		return "list[str]"
	case KindStringSliceSlice:
		return "list[list[str]]"
	case KindByteSliceSlice:
		return "list[list[str]]"
	case KindFloat64:
		return "float"
	case KindListNode:
		return "ListNode"
	case KindTreeNode:
		return "TreeNode"
	default:
		return "Any"
	}
}

func pyParam(params []ParamSpec) string {
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf("%s: %s", p.Name, pyType(p.Type))
	}
	return strings.Join(parts, ", ")
}

func pyReturn(ret ReturnSpec) string { return pyType(ret.Type) }

func pyCall(funcName string, params []ParamSpec) string {
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.Name
	}
	return funcName + "(" + strings.Join(parts, ", ") + ")"
}

func pyValue(v string, kind ValueKind) string {
	switch kind {
	case KindString:
		return `"` + strings.Trim(v, `"`) + `"`
	case KindStringSlice:
		return strings.ReplaceAll(strings.ReplaceAll(v, `"`, `'`), `[]string`, ``)
	case KindBool:
		if v == "true" {
			return "True"
		}
		return "False"
	default:
		return v
	}
}

func renderPythonStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("# %s\n# TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.PythonFuncName()
	buf.WriteString(fmt.Sprintf("def %s(%s) -> %s:\n", funcName, pyParam(spec.Params), pyReturn(spec.Return)))
	buf.WriteString("    # TODO: implement\n")
	if spec.Return.Type == KindInt {
		buf.WriteString("    return 0\n")
	} else if spec.Return.Type == KindBool {
		buf.WriteString("    return False\n")
	} else if spec.Return.Type == KindString {
		buf.WriteString("    return \"\"\n")
	} else {
		buf.WriteString("    pass\n")
	}
	return buf.Bytes()
}

func renderPythonTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("# %s\n# SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	funcName := spec.PythonFuncName()
	fileBase := spec.FileBase()
	buf.WriteString("import pytest\n")
	buf.WriteString(fmt.Sprintf("from %s import %s\n\n", fileBase, funcName))

	if spec.NeedsListNode {
		buf.WriteString("class ListNode:\n    def __init__(self, val=0, next=None):\n        self.val = val\n        self.next = next\n\n")
	}
	if spec.NeedsTreeNode {
		buf.WriteString("class TreeNode:\n    def __init__(self, val=0, left=None, right=None):\n        self.val = val\n        self.left = left\n        self.right = right\n\n")
	}

	if len(spec.Examples) == 1 {
		ex := spec.Examples[0]
		buf.WriteString(fmt.Sprintf("def test_%s():\n", funcName))
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		buf.WriteString(fmt.Sprintf("    assert %s == %s\n", pyCall(funcName, spec.Params), ex.Expect))
	} else {
		buf.WriteString("@pytest.mark.parametrize(\"")
		paramNames := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			paramNames[i] = p.Name
		}
		buf.WriteString(strings.Join(paramNames, ","))
		buf.WriteString(",expected\", [\n")
		for _, ex := range spec.Examples {
			buf.WriteString("    (")
			for i, p := range spec.Params {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(ex.Input[p.Name])
			}
			buf.WriteString(fmt.Sprintf(", %s),\n", ex.Expect))
		}
		buf.WriteString("])\n")
		buf.WriteString(fmt.Sprintf("def test_%s(%s, expected):\n", funcName, strings.Join(paramNames, ", ")))
		buf.WriteString(fmt.Sprintf("    assert %s == expected\n", pyCall(funcName, spec.Params)))
	}
	return buf.Bytes()
}
