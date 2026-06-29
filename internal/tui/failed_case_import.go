package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
)

func appendFailedLeetCodeCase(problem *roadmap.Problem, language string, testPath string, result *leetcode.SubmissionResult) (string, error) {
	if result == nil || strings.TrimSpace(result.LastTestcase) == "" || strings.TrimSpace(result.ExpectedOutput) == "" {
		return "", nil
	}
	if language != "go" {
		return "Remote failed testcase import is only supported for Go local tests for now.", nil
	}
	spec, ok := generator.SpecForProblem(problem)
	if !ok || spec.IsDesign || spec.Comparison == generator.CmpSkip || spec.Return.Type == generator.KindVoid {
		return "Remote failed testcase is not supported for this Problem shape yet.", nil
	}
	inputs := splitLeetCodeTestcase(result.LastTestcase)
	if len(inputs) != len(spec.Params) {
		return "Remote failed testcase did not match this Problem's parameter shape.", nil
	}

	parts := []string{`name: "leetcode failed"`}
	for i, p := range spec.Params {
		value, err := leetCodeLiteralToGo(inputs[i], p.Type)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s: %s", p.Name, value))
	}
	expect, err := leetCodeLiteralToGo(result.ExpectedOutput, spec.Return.Type)
	if err != nil {
		return "", err
	}
	parts = append(parts, "expect: "+expect)
	caseLine := "\t\t{" + strings.Join(parts, ", ") + "},"

	b, err := os.ReadFile(testPath)
	if err != nil {
		return "", err
	}
	content := string(b)
	if strings.Contains(content, caseLine) {
		return "Remote failed testcase already exists in local TestSuite.", nil
	}
	marker := "\t}\n\n\tfor _, tt := range tests {"
	if !strings.Contains(content, marker) {
		return "Generated Go TestSuite format is not supported for testcase import.", nil
	}
	updated := strings.Replace(content, marker, caseLine+"\n"+marker, 1)
	if err := os.WriteFile(testPath, []byte(updated), 0o644); err != nil {
		return "", err
	}
	return "Added remote failed testcase to local TestSuite.", nil
}

func failedLeetCodeExample(problem *roadmap.Problem, result *leetcode.SubmissionResult) (generator.ExampleSpec, error) {
	if result == nil || strings.TrimSpace(result.LastTestcase) == "" || strings.TrimSpace(result.ExpectedOutput) == "" {
		return generator.ExampleSpec{}, fmt.Errorf("missing failed LeetCode testcase")
	}
	spec, ok := generator.SpecForProblem(problem)
	if !ok || spec.IsDesign || spec.Comparison == generator.CmpSkip || spec.Return.Type == generator.KindVoid {
		return generator.ExampleSpec{}, fmt.Errorf("unsupported Problem shape")
	}
	inputs := splitLeetCodeTestcase(result.LastTestcase)
	if len(inputs) != len(spec.Params) {
		return generator.ExampleSpec{}, fmt.Errorf("failed testcase did not match parameter shape")
	}
	inputMap := map[string]string{"_name": "leetcode failed"}
	for i, p := range spec.Params {
		value, err := leetCodeLiteralToGo(inputs[i], p.Type)
		if err != nil {
			return generator.ExampleSpec{}, err
		}
		inputMap[p.Name] = value
	}
	expect, err := leetCodeLiteralToGo(result.ExpectedOutput, spec.Return.Type)
	if err != nil {
		return generator.ExampleSpec{}, err
	}
	return generator.ExampleSpec{Input: inputMap, Expect: expect}, nil
}

func appendScrapedLeetCodeExamples(problem *roadmap.Problem, language string, testPath string, examples []generator.ExampleSpec) (string, error) {
	if len(examples) == 0 {
		return "", nil
	}
	if language != "go" {
		return "Scraped testcase import is only supported for Go local tests for now.", nil
	}
	spec, ok := generator.SpecForProblem(problem)
	if !ok || spec.IsDesign || spec.Comparison == generator.CmpSkip || spec.Return.Type == generator.KindVoid {
		return "Scraped testcase import is not supported for this Problem shape yet.", nil
	}
	if _, err := os.Stat(testPath); err != nil {
		return "", nil
	}

	b, err := os.ReadFile(testPath)
	if err != nil {
		return "", err
	}
	content := string(b)
	marker := "\t}\n\n\tfor _, tt := range tests {"
	if !strings.Contains(content, marker) {
		return "Generated Go TestSuite format is not supported for scraped testcase import.", nil
	}
	existingSignatures := make(map[string]bool)
	if baseSpec, ok := generator.SpecForProblem(problem); ok {
		for _, ex := range baseSpec.Examples {
			if sig, err := exampleValueSignature(baseSpec, ex); err == nil {
				existingSignatures[sig] = true
			}
		}
	}
	var added []string
	for _, ex := range examples {
		sig, err := exampleValueSignature(spec, ex)
		if err != nil || existingSignatures[sig] {
			continue
		}
		line, err := goTestCaseLineFromExample(spec, ex)
		if err != nil {
			continue
		}
		if strings.Contains(content, line) {
			continue
		}
		existingSignatures[sig] = true
		added = append(added, line)
	}
	if len(added) == 0 {
		return "", nil
	}
	insert := strings.Join(added, "\n") + "\n" + marker
	updated := strings.Replace(content, marker, insert, 1)
	if err := os.WriteFile(testPath, []byte(updated), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Added %d scraped LeetCode testcase(s) to local TestSuite.", len(added)), nil
}

func splitLeetCodeTestcase(input string) []string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			values = append(values, line)
		}
	}
	return values
}

func leetCodeLiteralToGo(value string, kind generator.ValueKind) (string, error) {
	value = strings.TrimSpace(value)
	switch kind {
	case generator.KindInt, generator.KindFloat64:
		return value, nil
	case generator.KindBool:
		return strings.ToLower(value), nil
	case generator.KindString:
		return value, nil
	case generator.KindIntSlice:
		return goCompositeLiteral(value, "[]int"), nil
	case generator.KindIntSliceSlice:
		return goCompositeLiteral(value, "[][]int"), nil
	case generator.KindStringSlice:
		return goCompositeLiteral(value, "[]string"), nil
	case generator.KindStringSliceSlice:
		return goCompositeLiteral(value, "[][]string"), nil
	default:
		return "", fmt.Errorf("unsupported testcase value kind %s", kind)
	}
}

func goCompositeLiteral(value string, prefix string) string {
	value = strings.ReplaceAll(value, "[", "{")
	value = strings.ReplaceAll(value, "]", "}")
	return prefix + value
}
