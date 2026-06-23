package leetcode

import (
	"regexp"
	"strings"
	"unicode"
)

func adaptSubmissionCode(lang string, problemSlug string, code string) string {
	if lang != "golang" {
		return code
	}
	return adaptGoSubmissionCode(problemSlug, code)
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
	localName := slugToPascal(problemSlug)
	leetcodeName := slugToCamel(problemSlug)
	if localName == "" || leetcodeName == "" || localName == leetcodeName {
		return code
	}

	pattern := regexp.MustCompile(`\bfunc\s+` + regexp.QuoteMeta(localName) + `\s*\(`)
	return pattern.ReplaceAllString(code, "func "+leetcodeName+"(")
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
