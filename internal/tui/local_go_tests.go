package tui

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"

	"github.com/prettyletto/leetgo/internal/generator"
)

func localGoTestExamples(spec *generator.ProblemSpec, testPath string) ([]generator.ExampleSpec, error) {
	if spec == nil || spec.IsDesign || spec.Comparison == generator.CmpSkip || spec.Return.Type == generator.KindVoid {
		return nil, nil
	}
	if _, err := os.Stat(testPath); err != nil {
		return nil, nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testPath, nil, 0)
	if err != nil {
		return nil, err
	}

	var examples []generator.ExampleSpec
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		fields := keyedFields(lit)
		if !hasLocalGoTestFields(spec, fields) {
			return true
		}
		ex := generator.ExampleSpec{Input: map[string]string{}}
		if name, ok := stringLiteralValue(fields["name"]); ok {
			ex.Input["_name"] = name
		}
		for _, p := range spec.Params {
			value, err := goExprString(fset, fields[p.Name])
			if err != nil {
				return true
			}
			ex.Input[p.Name] = value
		}
		expect, err := goExprString(fset, fields["expect"])
		if err != nil {
			return true
		}
		ex.Expect = expect
		examples = append(examples, ex)
		return true
	})
	return examples, nil
}

func keyedFields(lit *ast.CompositeLit) map[string]ast.Expr {
	fields := make(map[string]ast.Expr)
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		fields[ident.Name] = kv.Value
	}
	return fields
}

func hasLocalGoTestFields(spec *generator.ProblemSpec, fields map[string]ast.Expr) bool {
	if fields["expect"] == nil {
		return false
	}
	for _, p := range spec.Params {
		if fields[p.Name] == nil {
			return false
		}
	}
	return true
}

func stringLiteralValue(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

func goExprString(fset *token.FileSet, expr ast.Expr) (string, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return "", err
	}
	return buf.String(), nil
}
