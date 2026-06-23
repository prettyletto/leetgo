package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type CppTemplate struct{}

func (c *CppTemplate) StubExt() string { return ".cpp" }
func (c *CppTemplate) TestExt() string { return "_test.cpp" }

func (c *CppTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderCppStub(spec), nil
}

func (c *CppTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderCppTest(spec), nil
}

func cppType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "int"
	case KindIntSlice:
		return "vector<int>"
	case KindIntSliceSlice:
		return "vector<vector<int>>"
	case KindBool:
		return "bool"
	case KindString:
		return "string"
	case KindStringSlice:
		return "vector<string>"
	case KindFloat64:
		return "double"
	case KindListNode:
		return "ListNode*"
	case KindTreeNode:
		return "TreeNode*"
	default:
		return "void"
	}
}

func renderCppStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	buf.WriteString("#include <vector>\n#include <string>\nusing namespace std;\n\n")
	if spec.NeedsListNode {
		buf.WriteString("struct ListNode { int val; ListNode *next; };\n\n")
	}
	if spec.NeedsTreeNode {
		buf.WriteString("struct TreeNode { int val; TreeNode *left; TreeNode *right; };\n\n")
	}
	buf.WriteString("class Solution {\npublic:\n")
	funcName := spec.CppFuncName()
	params := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		t := cppType(p.Type)
		if t == "vector<int>" || t == "vector<vector<int>>" || t == "string" {
			t = t + "&"
		}
		params[i] = fmt.Sprintf("%s %s", t, p.Name)
	}
	buf.WriteString(fmt.Sprintf("    %s %s(%s) {\n", cppType(spec.Return.Type), funcName, strings.Join(params, ", ")))
	buf.WriteString("        // TODO: implement\n")
	buf.WriteString("        return {};\n")
	buf.WriteString("    }\n")
	buf.WriteString("};\n")
	return buf.Bytes()
}

func renderCppTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	fileBase := spec.FileBase()
	buf.WriteString("#include <cassert>\n#include <vector>\n#include <string>\n\n")
	buf.WriteString(fmt.Sprintf("#include \"%s.cpp\"\n\n", fileBase))
	buf.WriteString("int main() {\n")
	buf.WriteString("    Solution solution;\n\n")
	for _, ex := range spec.Examples {
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		funcName := spec.CppFuncName()
		buf.WriteString(fmt.Sprintf("    assert(solution.%s(%s) == %s);\n", funcName, strings.Join(args, ", "), ex.Expect))
	}
	buf.WriteString("\n    return 0;\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}
