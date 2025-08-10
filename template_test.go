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

func TestExpandTemplateArray(t *testing.T) {
	templates := []string{
		"convert",
		"{{input}}",
		"-o",
		"{{output}}",
	}

	vars := map[string]string{
		"input":  "test.txt",
		"output": "test.png",
	}

	expected := []string{
		"convert",
		"test.txt",
		"-o",
		"test.png",
	}

	result, err := ExpandTemplateArray(templates, vars)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
		}
	}
}

func TestPrepareVariables(t *testing.T) {
	input := "test input"
	output := "test.png"
	lang := "go"

	vars := PrepareVariables(input, output, lang)

	if vars["input"] != input {
		t.Errorf("Expected input %s, got %s", input, vars["input"])
	}
	if vars["output"] != output {
		t.Errorf("Expected output %s, got %s", output, vars["output"])
	}
	if vars["lang"] != lang {
		t.Errorf("Expected lang %s, got %s", lang, vars["lang"])
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

func TestHasVariableInArray(t *testing.T) {
	templates := []string{
		"convert",
		"{{input}}",
		"-o",
		"output.png",
	}

	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"has_input", "input", true},
		{"no_output", "output", false},
		{"no_lang", "lang", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasVariableInArray(templates, tt.varName)
			if result != tt.expected {
				t.Errorf("Variable %s: expected %v, got %v", tt.varName, tt.expected, result)
			}
		})
	}
}
