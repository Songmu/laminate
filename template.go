package laminate

import (
	"regexp"
	"strings"
)

var templateVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9]+)\s*\}\}`)

// ExpandTemplate expands template variables in a command string
func ExpandTemplate(template string, vars map[string]string) (string, error) {
	result := templateVarPattern.ReplaceAllStringFunc(template, func(match string) string {
		varName := strings.TrimSpace(match[2 : len(match)-2]) // Remove '{{' and '}}'
		if value, ok := vars[varName]; ok {
			return value
		}
		return match
	})

	return result, nil
}
