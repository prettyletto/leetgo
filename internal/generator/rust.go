package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type RustTemplate struct{}

func (r *RustTemplate) StubExt() string { return ".rs" }
func (r *RustTemplate) TestExt() string { return "_test.rs" }

func (r *RustTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderRustStub(spec), nil
}

func (r *RustTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d", p.ID)
	}
	return renderRustTest(spec), nil
}

func rustType(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "i32"
	case KindIntSlice:
		return "Vec<i32>"
	case KindIntSliceSlice:
		return "Vec<Vec<i32>>"
	case KindBool:
		return "bool"
	case KindString:
		return "String"
	case KindStringSlice:
		return "Vec<String>"
	case KindFloat64:
		return "f64"
	default:
		return "()"
	}
}

func rustZero(kind ValueKind) string {
	switch kind {
	case KindInt:
		return "0"
	case KindBool:
		return "false"
	case KindString:
		return "String::new()"
	case KindIntSlice:
		return "vec![]"
	case KindFloat64:
		return "0.0"
	default:
		return "panic!(\"not implemented\")"
	}
}

func renderRustStub(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// TODO: implement design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	buf.WriteString("pub struct Solution;\n\n")
	buf.WriteString("impl Solution {\n")
	funcName := spec.RustFuncName()
	params := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		params[i] = fmt.Sprintf("%s: %s", p.Name, rustType(p.Type))
	}
	buf.WriteString(fmt.Sprintf("    pub fn %s(%s) -> %s {\n", funcName, strings.Join(params, ", "), rustType(spec.Return.Type)))
	buf.WriteString("        // TODO: implement\n")
	buf.WriteString(fmt.Sprintf("        %s\n", rustZero(spec.Return.Type)))
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func renderRustTest(spec *ProblemSpec) []byte {
	var buf bytes.Buffer
	if spec.IsDesign {
		buf.WriteString(fmt.Sprintf("// %s\n// SKIPPED: design problem\n", spec.DesignNote))
		return buf.Bytes()
	}
	fileBase := spec.FileBase()
	buf.WriteString(fmt.Sprintf("#[path = \"%s.rs\"]\nmod %s;\n\nuse %s::Solution;\n\n", fileBase, fileBase, fileBase))
	funcName := spec.RustFuncName()
	buf.WriteString(fmt.Sprintf("#[test]\nfn test_%s() {\n", funcName))
	for _, ex := range spec.Examples {
		args := make([]string, len(spec.Params))
		for i, p := range spec.Params {
			args[i] = ex.Input[p.Name]
		}
		buf.WriteString(fmt.Sprintf("    assert_eq!(Solution::%s(%s), %s);\n", funcName, strings.Join(args, ", "), ex.Expect))
	}
	buf.WriteString("}\n")
	return buf.Bytes()
}
