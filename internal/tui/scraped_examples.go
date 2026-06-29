package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prettyletto/leetgo/internal/generator"
)

var exampleHeadingPattern = regexp.MustCompile(`(?i)^example\s+\d+\s*:`)

func scrapedExamplesFromDescription(spec *generator.ProblemSpec, description string) ([]generator.ExampleSpec, error) {
	blocks := leetcodeExampleBlocks(description)
	examples := make([]generator.ExampleSpec, 0, len(blocks))
	for i, block := range blocks {
		ex, err := scrapedExampleFromBlock(spec, block, i)
		if err != nil {
			continue
		}
		examples = append(examples, ex)
	}
	if len(examples) == 0 {
		return nil, fmt.Errorf("no LeetCode examples parsed")
	}
	return examples, nil
}

func leetcodeExampleBlocks(description string) []string {
	lines := strings.Split(description, "\n")
	var blocks []string
	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if exampleHeadingPattern.MatchString(trimmed) {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
			}
			current = []string{trimmed}
			continue
		}
		if len(current) > 0 {
			if strings.EqualFold(trimmed, "Constraints:") || strings.HasPrefix(strings.ToLower(trimmed), "constraints:") {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
				continue
			}
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func scrapedExampleFromBlock(spec *generator.ProblemSpec, block string, index int) (generator.ExampleSpec, error) {
	input, output, err := exampleInputOutput(block)
	if err != nil {
		return generator.ExampleSpec{}, err
	}
	assignments, err := parseLeetCodeAssignments(input)
	if err != nil {
		return generator.ExampleSpec{}, err
	}
	inputMap := map[string]string{"_name": fmt.Sprintf("leetcode example %d", index+1)}
	for _, p := range spec.Params {
		value, ok := assignments[p.Name]
		if !ok {
			return generator.ExampleSpec{}, fmt.Errorf("missing input %q", p.Name)
		}
		goValue, err := leetCodeLiteralToGo(value, p.Type)
		if err != nil {
			return generator.ExampleSpec{}, err
		}
		inputMap[p.Name] = goValue
	}
	expect, err := leetCodeLiteralToGo(output, spec.Return.Type)
	if err != nil {
		return generator.ExampleSpec{}, err
	}
	return generator.ExampleSpec{Input: inputMap, Expect: expect}, nil
}

func exampleInputOutput(block string) (string, string, error) {
	lines := strings.Split(block, "\n")
	var inputLines []string
	var outputLines []string
	section := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "input:"):
			section = "input"
			inputLines = append(inputLines, strings.TrimSpace(trimmed[len("Input:"):]))
		case strings.HasPrefix(lower, "output:"):
			section = "output"
			outputLines = append(outputLines, strings.TrimSpace(trimmed[len("Output:"):]))
		case strings.HasPrefix(lower, "explanation:"):
			section = ""
		case section == "input" && trimmed != "":
			inputLines = append(inputLines, trimmed)
		case section == "output" && trimmed != "":
			outputLines = append(outputLines, trimmed)
		}
	}
	input := strings.TrimSpace(strings.Join(inputLines, " "))
	output := strings.TrimSpace(strings.Join(outputLines, " "))
	if input == "" || output == "" {
		return "", "", fmt.Errorf("missing input or output")
	}
	return input, output, nil
}

func parseLeetCodeAssignments(input string) (map[string]string, error) {
	parts := splitTopLevel(input, ',')
	assignments := make(map[string]string, len(parts))
	for _, part := range parts {
		name, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid assignment %q", part)
		}
		assignments[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return assignments, nil
}

func splitTopLevel(input string, sep rune) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	inString := false
	escaped := false
	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && inString {
			current.WriteRune(r)
			escaped = true
			continue
		}
		if r == '"' {
			inString = !inString
			current.WriteRune(r)
			continue
		}
		if !inString {
			switch r {
			case '[', '{', '(':
				depth++
			case ']', '}', ')':
				if depth > 0 {
					depth--
				}
			case sep:
				if depth == 0 {
					parts = append(parts, strings.TrimSpace(current.String()))
					current.Reset()
					continue
				}
			}
		}
		current.WriteRune(r)
	}
	if strings.TrimSpace(current.String()) != "" {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

func goTestCaseLineFromExample(spec *generator.ProblemSpec, ex generator.ExampleSpec) (string, error) {
	name := ex.Input["_name"]
	if name == "" {
		name = "leetcode example"
	}
	quotedName := strconv.Quote(name)
	parts := []string{"name: " + quotedName}
	for _, p := range spec.Params {
		value := ex.Input[p.Name]
		if value == "" {
			return "", fmt.Errorf("missing input %q", p.Name)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", p.Name, value))
	}
	if ex.Expect == "" {
		return "", fmt.Errorf("missing expected output")
	}
	parts = append(parts, "expect: "+ex.Expect)
	return "\t\t{" + strings.Join(parts, ", ") + "},", nil
}

func exampleValueSignature(spec *generator.ProblemSpec, ex generator.ExampleSpec) (string, error) {
	parts := make([]string, 0, len(spec.Params)+1)
	for _, p := range spec.Params {
		value := ex.Input[p.Name]
		if value == "" {
			return "", fmt.Errorf("missing input %q", p.Name)
		}
		parts = append(parts, p.Name+"="+normalizeGoLiteral(value))
	}
	if ex.Expect == "" {
		return "", fmt.Errorf("missing expected output")
	}
	parts = append(parts, "expect="+normalizeGoLiteral(ex.Expect))
	return strings.Join(parts, "|"), nil
}

func normalizeGoLiteral(value string) string {
	return strings.Join(strings.Fields(value), "")
}
