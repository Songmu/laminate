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
