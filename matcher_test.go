package laminate

import (
	"testing"
)

func TestMatchLanguage(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		lang     string
		expected bool
		hasError bool
	}{
		{"exact_match", "go", "go", true, false},
		{"no_match", "go", "python", false, false},
		{"wildcard", "*", "anything", true, false},
		{"glob_pattern", "go*", "golang", true, false},
		{"glob_pattern_no_match", "go*", "python", false, false},
		{"brace_expansion", "{go,python,rust}", "go", true, false},
		{"brace_expansion_match", "{go,python,rust}", "python", true, false},
		{"brace_expansion_no_match", "{go,python,rust}", "java", false, false},
		{"complex_brace", "{c,cpp,c++}", "cpp", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matchLanguage(tt.pattern, tt.lang)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Pattern %s, Lang %s: expected %v, got %v", tt.pattern, tt.lang, tt.expected, result)
				}
			}
		})
	}
}

func TestFindMatchingCommand(t *testing.T) {
	commands := []*Command{
		{Lang: "go", Run: RunCommand{str: "cmd1"}, Ext: "png"},
		{Lang: "{python,py}", Run: RunCommand{str: "cmd2"}, Ext: "jpg"},
		{Lang: "*", Run: RunCommand{str: "cmd3"}, Ext: "gif"},
	}

	tests := []struct {
		name        string
		lang        string
		expectedCmd string
		hasError    bool
	}{
		{"go_match", "go", "cmd1", false},
		{"python_brace_match", "python", "cmd2", false},
		{"py_brace_match", "py", "cmd2", false},
		{"wildcard_match", "unknown", "cmd3", false},
		{"first_match_wins", "go", "cmd1", false}, // Should match first, not wildcard
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := FindMatchingCommand(commands, tt.lang)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if cmd == nil {
					t.Error("Expected command, got nil")
				} else if cmd.Run.String() != tt.expectedCmd {
					t.Errorf("Expected command %s, got %s", tt.expectedCmd, cmd.Run.String())
				}
			}
		})
	}
}
