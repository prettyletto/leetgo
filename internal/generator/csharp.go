package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type CSharpTemplate struct{}

func (c *CSharpTemplate) StubExt() string { return ".cs" }
func (c *CSharpTemplate) TestExt() string { return "Tests.cs" }

func (c *CSharpTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderCSharpStub(spec), nil
}

func (c *CSharpTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderCSharpTest(spec), nil
}

func csType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "int"
	case KindIntSlice:
		return "int[]"
	case KindIntSliceSlice:
		return "int[][]"
	case KindBool:
		return "bool"
	case KindString:
		return "string"
	case KindStringSlice:
		return "string[]"
	case KindFloat64:
		return "double"
	case KindListNode:
		return "ListNode"
	case KindTreeNode:
		return "TreeNode"
	default:
		return "void"
	}
}

func csZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindBool:
		return "false"
	case KindString:
		return "\"\""
	case KindIntSlice:
		return "new int[] {}"
	default:
		return "null"
	}
}

func renderCSharpStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	if spec.NeedsListNode {
		buf.WriteString(nodeStructCS)
	}
	if spec.NeedsTreeNode {
		buf.WriteString(treeStructCS)
	}
	buf.WriteString("public class Solution {\n")
	funcName := spec.CSharpFuncName()
	params := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		params[i] = fmt.Sprintf("%s %s", csType(p.Type), p.Name)
	}
	buf.WriteString(fmt.Sprintf("    public %s %s(%s) {\n", csType(spec.Return.Type), funcName, strings.Join(params, ", ")))
	buf.WriteString("        // TODO: implement\n")
	buf.WriteString(fmt.Sprintf("        return %s;\n", csZero(spec.Return.Type)))
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func renderCSharpTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	buf.WriteString("using Xunit;\n\n")
	if spec.NeedsListNode {
		buf.WriteString(nodeStructCS)
	}
	if spec.NeedsTreeNode {
		buf.WriteString(treeStructCS)
	}
	buf.WriteString("public class SolutionTests {\n")
	buf.WriteString("    [Fact]\n")
	funcName := spec.CSharpFuncName()
	buf.WriteString(fmt.Sprintf("    public void %s_PassesExamples() {\n", funcName))
	buf.WriteString("        Solution solution = new Solution();\n\n")
	for _, ex := range spec.Examples {
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		buf.WriteString(fmt.Sprintf("        Assert.Equal(%s, solution.%s(%s));\n", ex.Expect, funcName, strings.Join(args, ", ")))
	}
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

const nodeStructCS = `
public class ListNode {
    public int val;
    public ListNode next;
    public ListNode(int val = 0, ListNode next = null) { this.val = val; this.next = next; }
}

`

const treeStructCS = `
public class TreeNode {
    public int val;
    public TreeNode left;
    public TreeNode right;
    public TreeNode(int val = 0, TreeNode left = null, TreeNode right = null) { this.val = val; this.left = left; this.right = right; }
}

`
