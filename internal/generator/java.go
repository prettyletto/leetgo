package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type JavaTemplate struct{}

func (j *JavaTemplate) StubExt() string { return ".java" }
func (j *JavaTemplate) TestExt() string { return "Test.java" }

func (j *JavaTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderJavaStub(spec), nil
}

func (j *JavaTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderJavaTest(spec), nil
}

func javaType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "int"
	case KindIntSlice:
		return "int[]"
	case KindIntSliceSlice:
		return "int[][]"
	case KindBool:
		return "boolean"
	case KindString:
		return "String"
	case KindStringSlice:
		return "List<String>"
	case KindFloat64:
		return "double"
	case KindListNode:
		return "ListNode"
	case KindTreeNode:
		return "TreeNode"
	default:
		return "Object"
	}
}

func javaZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindBool:
		return "false"
	case KindIntSlice:
		return "new int[] {}"
	case KindString:
		return "\"\""
	default:
		return "null"
	}
}

func renderJavaStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	if spec.NeedsListNode {
		buf.WriteString(nodeStructJava)
	}
	if spec.NeedsTreeNode {
		buf.WriteString(treeStructJava)
	}
	buf.WriteString("class Solution {\n")
	funcName := spec.JavaFuncName()
	params := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		params[i] = fmt.Sprintf("%s %s", javaType(p.Type), p.Name)
	}
	buf.WriteString(fmt.Sprintf("    public %s %s(%s) {\n", javaType(spec.Return.Type), funcName, strings.Join(params, ", ")))
	buf.WriteString("        // TODO: implement\n")
	buf.WriteString(fmt.Sprintf("        return %s;\n", javaZero(spec.Return.Type)))
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func renderJavaTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	buf.WriteString("import org.junit.jupiter.api.Test;\n")
	buf.WriteString("import static org.junit.jupiter.api.Assertions.*;\n\n")
	if spec.NeedsListNode {
		buf.WriteString(nodeStructJava)
	}
	if spec.NeedsTreeNode {
		buf.WriteString(treeStructJava)
	}
	buf.WriteString("class SolutionTest {\n")
	buf.WriteString("    @Test\n")
	funcName := spec.JavaFuncName()
	buf.WriteString(fmt.Sprintf("    void test%s() {\n", funcName))
	buf.WriteString("        Solution solution = new Solution();\n")
	for _, ex := range spec.Examples {
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		buf.WriteString(fmt.Sprintf("        assertArrayEquals(%s, solution.%s(%s));\n", ex.Expect, funcName, strings.Join(args, ", ")))
	}
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

const nodeStructJava = `
class ListNode {
    int val;
    ListNode next;
    ListNode() {}
    ListNode(int val) { this.val = val; }
    ListNode(int val, ListNode next) { this.val = val; this.next = next; }
}

`

const treeStructJava = `
class TreeNode {
    int val;
    TreeNode left;
    TreeNode right;
    TreeNode() {}
    TreeNode(int val) { this.val = val; }
    TreeNode(int val, TreeNode left, TreeNode right) { this.val = val; this.left = left; this.right = right; }
}

`
