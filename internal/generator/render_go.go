package generator

import (
	"bytes"
	"fmt"
	"strings"
)

func goType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "int"
	case KindIntSlice:
		return "[]int"
	case KindIntSliceSlice:
		return "[][]int"
	case KindBool:
		return "bool"
	case KindString:
		return "string"
	case KindStringSlice:
		return "[]string"
	case KindStringSliceSlice:
		return "[][]string"
	case KindByteSlice:
		return "[]byte"
	case KindByteSliceSlice:
		return "[][]byte"
	case KindFloat64:
		return "float64"
	case KindListNode:
		return "*ListNode"
	case KindTreeNode:
		return "*TreeNode"
	default:
		return "interface{}"
	}
}

func goZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindIntSlice:
		return "nil"
	case KindIntSliceSlice:
		return "nil"
	case KindBool:
		return "false"
	case KindString:
		return "\"\""
	case KindStringSlice:
		return "nil"
	case KindStringSliceSlice:
		return "nil"
	case KindByteSlice:
		return "nil"
	case KindByteSliceSlice:
		return "nil"
	case KindFloat64:
		return "0.0"
	case KindListNode:
		return "nil"
	case KindTreeNode:
		return "nil"
	default:
		return "nil"
	}
}

func goParam(params []ParamSpec) string {
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.Name + " " + goType(p.Type)
	}
	return strings.Join(parts, ", ")
}

func goTestFuncName(funcName string) string {
	if funcName == "" {
		return "Solution"
	}
	runes := []rune(funcName)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func goReturn(ret ReturnSpec) string {
	return goType(ret.Type)
}

func needsDeepEqual(kind ValueKind) bool {
	switch kind {
	case KindIntSlice, KindIntSliceSlice, KindStringSlice, KindStringSliceSlice, KindByteSlice, KindByteSliceSlice:
		return true
	case KindListNode, KindTreeNode:
		return true
	default:
		return false
	}
}

func renderGoStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer

	buf.WriteString("package solution\n\n")

	if spec.IsDesign {
		buf.WriteString("// " + spec.DesignNote + "\n")
		buf.WriteString("// TODO: implement design problem\n")
		return buf.Bytes()
	}

	if spec.NeedsListNode {
		buf.WriteString(nodeStructGo)
	}

	if spec.NeedsTreeNode {
		buf.WriteString(treeStructGo)
	}

	funcName := spec.GoFuncName()
	retType := goReturn(spec.Return)
	buf.WriteString(fmt.Sprintf("func %s(%s) %s {\n", funcName, goParam(spec.Params), retType))
	buf.WriteString("\t// TODO: implement\n")
	buf.WriteString(fmt.Sprintf("\treturn %s\n", goZero(spec.Return.Type)))
	buf.WriteString("}\n")

	return buf.Bytes()
}

func renderGoTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer

	buf.WriteString("package solution\n\n")

	if spec.IsDesign {
		buf.WriteString("// " + spec.DesignNote + "\n")
		buf.WriteString("// SKIPPED: design problems require class/struct generation.\n")
		return buf.Bytes()
	}

	if spec.Comparison == CmpSkip {
		buf.WriteString("import \"testing\"\n\n")
		funcName := spec.GoFuncName()
		buf.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", goTestFuncName(funcName)))
		buf.WriteString("\tt.Skip(\"comparison not yet supported for this problem shape\")\n")
		buf.WriteString("}\n")
		return buf.Bytes()
	}

	buf.WriteString("import (\n")
	needsReflect := needsDeepEqual(spec.Return.Type) || spec.Comparison == CmpDeep || spec.Comparison == CmpUnordered || spec.Return.Type == KindVoid
	if needsReflect {
		buf.WriteString("\t\"reflect\"\n")
	}
	buf.WriteString("\t\"testing\"\n")
	buf.WriteString(")\n\n")

	if spec.Comparison == CmpUnordered {
		buf.WriteString("\nfunc unorderedEqual(a, b [][]string) bool {\n")
		buf.WriteString("\tif len(a) != len(b) { return false }\n")
		buf.WriteString("\tused := make([]bool, len(b))\n")
		buf.WriteString("\tfor _, x := range a {\n")
		buf.WriteString("\t\tfound := false\n")
		buf.WriteString("\t\tfor j, y := range b {\n")
		buf.WriteString("\t\t\tif !used[j] && reflect.DeepEqual(x, y) { used[j] = true; found = true; break }\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t\tif !found { return false }\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn true\n")
		buf.WriteString("}\n\n")
	}

	funcName := spec.GoFuncName()

	if spec.Return.Type == KindVoid {
		buf.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", goTestFuncName(funcName)))
		for _, ex := range spec.Examples {
			firstParam := spec.Params[0].Name
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", firstParam, ex.Input[firstParam]))
			callArgs := make([]string, len(spec.Params))
			for i, p := range spec.Params {
				callArgs[i] = p.Name
			}
			buf.WriteString(fmt.Sprintf("\t%s(%s)\n", funcName, strings.Join(callArgs, ", ")))
			buf.WriteString(fmt.Sprintf("\twant := %s\n", ex.Expect))
			buf.WriteString(fmt.Sprintf("\tif !reflect.DeepEqual(%s, want) {\n", firstParam))
			buf.WriteString(fmt.Sprintf("\t\tt.Errorf(\"got %%v, want %%v\", %s, want)\n", firstParam))
			buf.WriteString("\t}\n")
		}
		buf.WriteString("}\n")
		return buf.Bytes()
	}

	buf.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", goTestFuncName(funcName)))

	buf.WriteString("\ttests := []struct {\n")
	buf.WriteString("\t\tname string\n")
	for _, p := range spec.Params {
		buf.WriteString(fmt.Sprintf("\t\t%s %s\n", p.Name, goType(p.Type)))
	}
	buf.WriteString(fmt.Sprintf("\t\texpect %s\n", goReturn(spec.Return)))
	buf.WriteString("\t}{\n")
	for _, ex := range spec.Examples {
		name := ex.Input["_name"]
		buf.WriteString(fmt.Sprintf("\t\t{name: \"%s\", ", name))
		for _, p := range spec.Params {
			buf.WriteString(fmt.Sprintf("%s: %s, ", p.Name, ex.Input[p.Name]))
		}
		buf.WriteString(fmt.Sprintf("expect: %s},\n", ex.Expect))
	}
	buf.WriteString("\t}\n\n")

	buf.WriteString("\tfor _, tt := range tests {\n")
	buf.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	callArgs := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		callArgs[i] = "tt." + p.Name
	}
	buf.WriteString(fmt.Sprintf("\t\t\tgot := %s(%s)\n", funcName, strings.Join(callArgs, ", ")))

	useDeep := needsDeepEqual(spec.Return.Type) || spec.Comparison == CmpDeep || spec.Comparison == CmpUnordered
	if useDeep {
		buf.WriteString("\t\t\tif !reflect.DeepEqual(got, tt.expect) {\n")
	} else {
		buf.WriteString("\t\t\tif got != tt.expect {\n")
	}
	buf.WriteString("\t\t\t\tt.Errorf(\"got %v, want %v\", got, tt.expect)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t})\n")
	buf.WriteString("\t}\n")

	buf.WriteString("}\n")
	return buf.Bytes()
}

const nodeStructGo = `
type ListNode struct {
	Val  int
	Next *ListNode
}

`

const treeStructGo = `
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

`
