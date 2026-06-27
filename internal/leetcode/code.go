package leetcode

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/prettyletto/leetgo/internal/generator"
)

func adaptSubmissionCode(lang string, problemSlug string, code string) string {
	switch lang {
	case "golang":
		return adaptGoSubmissionCode(problemSlug, code)
	case "python3":
		return adaptPythonSubmissionCode(problemSlug, code)
	case "typescript":
		return adaptTypeScriptSubmissionCode(code)
	case "javascript":
		return adaptJavaScriptSubmissionCode(code)
	default:
		return code
	}
}

func adaptGoSubmissionCode(problemSlug string, code string) string {
	lines := strings.Split(code, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "package ") {
		lines = lines[1:]
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
	}

	code = strings.Join(lines, "\n")
	leetcodeName := goLeetCodeFuncName(problemSlug)
	if leetcodeName == "" {
		return code
	}

	for _, localName := range goLocalSubmissionNames(problemSlug, leetcodeName) {
		if localName == "" || localName == leetcodeName {
			continue
		}
		pattern := regexp.MustCompile(`\bfunc\s+` + regexp.QuoteMeta(localName) + `\s*\(`)
		code = pattern.ReplaceAllString(code, "func "+leetcodeName+"(")
	}
	return code
}

func adaptPythonSubmissionCode(problemSlug string, code string) string {
	funcName := goLeetCodeFuncName(problemSlug)
	if funcName == "" || strings.Contains(code, "class Solution") {
		return code
	}

	pattern := regexp.MustCompile(`^def\s+` + regexp.QuoteMeta(funcName) + `\s*\(([^)]*)\)(\s*->\s*[^:]+)?\s*:`)
	lines := strings.Split(code, "\n")
	wrapped := make([]string, 0, len(lines)+1)
	inSolution := false
	found := false
	for _, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if matches != nil {
			params := strings.TrimSpace(matches[1])
			if params == "" {
				params = "self"
			} else if !strings.HasPrefix(params, "self") {
				params = "self, " + params
			}
			wrapped = append(wrapped, "class Solution:")
			wrapped = append(wrapped, "    def "+funcName+"("+params+")"+matches[2]+":")
			inSolution = true
			found = true
			continue
		}
		if inSolution && line != "" {
			wrapped = append(wrapped, "    "+line)
			continue
		}
		wrapped = append(wrapped, line)
	}
	if !found {
		return code
	}
	return strings.Join(wrapped, "\n")
}

func adaptTypeScriptSubmissionCode(code string) string {
	code = regexp.MustCompile(`(?m)^export\s+function\s+`).ReplaceAllString(code, "function ")
	code = regexp.MustCompile(`(?m)^export\s+`).ReplaceAllString(code, "")
	return code
}

func adaptJavaScriptSubmissionCode(code string) string {
	return regexp.MustCompile(`(?m)^\s*module\.exports\s*=\s*\w+\s*;\s*$\n?`).ReplaceAllString(code, "")
}

func goLeetCodeFuncName(problemSlug string) string {
	if spec, ok := generator.SpecForSlug(problemSlug); ok {
		return spec.GoFuncName()
	}
	return slugToCamel(problemSlug)
}

func goLocalSubmissionNames(problemSlug, leetcodeName string) []string {
	names := []string{slugToPascal(problemSlug), slugToCamel(problemSlug)}
	if spec, ok := generator.SpecForSlug(problemSlug); ok {
		names = append(names, spec.GoFuncName(), spec.FuncName)
	}
	names = append(names, leetcodeName)
	return names
}

func slugToPascal(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool { return r == '-' || r == '_' || unicode.IsSpace(r) })
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func slugToCamel(slug string) string {
	pascal := slugToPascal(slug)
	if pascal == "" {
		return ""
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
