package laminate

import (
	"fmt"
	"regexp"
)

var templateVarPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// ExpandTemplate expands template variables in a command string
func ExpandTemplate(template string, vars map[string]string) (string, error) {
	result := templateVarPattern.ReplaceAllStringFunc(template, func(match string) string {
		varName := match[2 : len(match)-2] // Remove {{ and }}
		if value, ok := vars[varName]; ok {
			return value
		}
		return match
	})

	return result, nil
}

// ExpandTemplateArray expands template variables in a command array
func ExpandTemplateArray(templates []string, vars map[string]string) ([]string, error) {
	var result []string
	for _, template := range templates {
		expanded, err := ExpandTemplate(template, vars)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded)
	}
	return result, nil
}

// PrepareVariables prepares the template variables for expansion
func PrepareVariables(input, output, lang string) map[string]string {
	return map[string]string{
		"input":  input,
		"output": output,
		"lang":   lang,
	}
}

// HasVariable checks if a template contains a specific variable
func HasVariable(template string, varName string) bool {
	pattern := fmt.Sprintf(`\{\{%s\}\}`, regexp.QuoteMeta(varName))
	matched, _ := regexp.MatchString(pattern, template)
	return matched
}

// HasVariableInArray checks if any template in array contains a specific variable
func HasVariableInArray(templates []string, varName string) bool {
	for _, template := range templates {
		if HasVariable(template, varName) {
			return true
		}
	}
	return false
}
