package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prettyletto/leetgo/internal/generator"
)

const leetgoDebugTestName = "TestLeetgoDebugCase"

func writeGoDebugTestFile(dir string, spec *generator.ProblemSpec, exampleIndex int) (string, error) {
	if exampleIndex < 0 || exampleIndex >= len(spec.Examples) {
		return "", fmt.Errorf("debug case index out of range")
	}
	if spec.IsDesign || spec.Comparison == generator.CmpSkip || spec.Return.Type == generator.KindVoid {
		return "", fmt.Errorf("this Problem shape is not supported by Go debug test generation yet")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "leetgo_debug_test.go")
	content, err := renderGoDebugTest(spec, exampleIndex)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func renderGoDebugTest(spec *generator.ProblemSpec, exampleIndex int) ([]byte, error) {
	ex := spec.Examples[exampleIndex]
	var buf bytes.Buffer
	buf.WriteString("package solution\n\n")
	buf.WriteString("import (\n")
	if goDebugNeedsReflect(spec) {
		buf.WriteString("\t\"reflect\"\n")
	}
	buf.WriteString("\t\"testing\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", leetgoDebugTestName))
	for _, p := range spec.Params {
		value := ex.Input[p.Name]
		if value == "" {
			return nil, fmt.Errorf("missing debug case input %q", p.Name)
		}
		buf.WriteString(fmt.Sprintf("\t%s := %s\n", p.Name, value))
	}
	args := make([]string, len(spec.Params))
	for i, p := range spec.Params {
		args[i] = p.Name
	}
	buf.WriteString(fmt.Sprintf("\tgot := %s(%s)\n", spec.GoFuncName(), strings.Join(args, ", ")))
	buf.WriteString(fmt.Sprintf("\texpect := %s\n", ex.Expect))
	if goDebugNeedsReflect(spec) {
		buf.WriteString("\tif !reflect.DeepEqual(got, expect) {\n")
	} else {
		buf.WriteString("\tif got != expect {\n")
	}
	buf.WriteString("\t\tt.Fatalf(\"got %v, want %v\", got, expect)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

func goDebugNeedsReflect(spec *generator.ProblemSpec) bool {
	if spec.Comparison == generator.CmpDeep || spec.Comparison == generator.CmpUnordered {
		return true
	}
	switch spec.Return.Type {
	case generator.KindIntSlice, generator.KindIntSliceSlice, generator.KindStringSlice, generator.KindStringSliceSlice, generator.KindByteSlice, generator.KindByteSliceSlice, generator.KindListNode, generator.KindTreeNode:
		return true
	default:
		return false
	}
}
