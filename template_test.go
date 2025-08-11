package laminate

import (
	"testing"
)

func TestExpandTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		expected string
	}{
		{
			"no_variables",
			"echo hello",
			map[string]string{},
			"echo hello",
		},
		{
			"single_variable",
			"echo {{input}}",
			map[string]string{"input": "world"},
			"echo world",
		},
		{
			"multiple_variables",
			"convert {{input}} -o {{output}}",
			map[string]string{"input": "test.txt", "output": "test.png"},
			"convert test.txt -o test.png",
		},
		{
			"missing_variable",
			"echo {{input}} {{missing}}",
			map[string]string{"input": "hello"},
			"echo hello {{missing}}",
		},
		{
			"repeated_variable",
			"echo {{input}} and {{input}} again",
			map[string]string{"input": "test"},
			"echo test and test again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandTemplate(tt.template, tt.vars)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHasVariable(t *testing.T) {
	tests := []struct {
		name     string
		template string
		varName  string
		expected bool
	}{
		{"has_variable", "echo {{input}}", "input", true},
		{"no_variable", "echo hello", "input", false},
		{"multiple_variables", "convert {{input}} -o {{output}}", "output", true},
		{"partial_match", "echo input", "input", false},
		{"special_chars", "echo {{var.name}}", "var.name", true},
		{"escaped_special_chars", "echo {{var[0]}}", "var[0]", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasVariable(tt.template, tt.varName)
			if result != tt.expected {
				t.Errorf("Template %s, Variable %s: expected %v, got %v", tt.template, tt.varName, tt.expected, result)
			}
		})
	}
}
