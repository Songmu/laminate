package laminate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

func TestCommand_GetExt(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{"default", "", "png"},
		{"custom", "jpg", "jpg"},
		{"with_dot", ".gif", ".gif"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Ext: tt.ext}
			result := cmd.GetExt()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRunCommand_UnmarshalYAML_WithActualYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		isArray  bool
		expected any
	}{
		{
			"string_command",
			`run: "echo hello"`,
			false,
			"echo hello",
		},
		{
			"array_command",
			`run: ["echo", "hello"]`,
			true,
			[]string{"echo", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd struct {
				Run RunCommand `yaml:"run"`
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &cmd)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if cmd.Run.IsArray() != tt.isArray {
				t.Errorf("Expected isArray=%v, got %v", tt.isArray, cmd.Run.IsArray())
			}

			if tt.isArray {
				result := cmd.Run.Array()
				expected := tt.expected.([]string)
				if len(result) != len(expected) {
					t.Errorf("Expected %v, got %v", expected, result)
				}
				for i, v := range result {
					if v != expected[i] {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			} else {
				result := cmd.Run.String()
				if result != tt.expected.(string) {
					t.Errorf("Expected %s, got %s", tt.expected.(string), result)
				}
			}
		})
	}
}

func TestPathOverride(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		getFunc  func() string
	}{
		{
			name:     "config_path_override",
			envVar:   "LAMINATE_CONFIG_PATH",
			envValue: "/tmp/test-config.yaml",
			getFunc:  getConfigPath,
		},
		{
			name:     "cache_path_override",
			envVar:   "LAMINATE_CACHE_PATH",
			envValue: "/tmp/test-cache",
			getFunc:  getCachePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envValue)

			result := tt.getFunc()
			if result != tt.envValue {
				t.Errorf("Expected %s, got %s", tt.envValue, result)
			}
		})
	}
}

func mustParseDuration(value string) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		panic("Invalid duration: " + value)
	}
	return d
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedCache  time.Duration
		expectedCmdLen int
		expectedLang   string
		expectedExt    string
		expectError    bool
	}{
		{
			name:           "file_not_exists",
			expectedCmdLen: 0,
			expectError:    false,
		},
		{
			name: "valid_file",
			configContent: `cache: 1h
commands:
- lang: test
  run: echo test
  ext: png
`,
			expectedCache:  mustParseDuration("1h"),
			expectedCmdLen: 1,
			expectedLang:   "test",
			expectedExt:    "png",
			expectError:    false,
		},
		{
			name: "valid_file_without_cache",
			configContent: `commands:
- lang: test
  run: echo test
  ext: png
`,
			expectedCmdLen: 1,
			expectedLang:   "test",
			expectedExt:    "png",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configContent != "" {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")

				err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
				defer os.RemoveAll(tmpDir)
				t.Setenv("LAMINATE_CONFIG_PATH", configFile)
			} else {
				t.Setenv("LAMINATE_CONFIG_PATH", "/non/existent/config.yaml")
			}

			config, err := LoadConfig()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if config == nil {
				t.Error("Expected config, got nil")
				return
			}

			if config.Cache != tt.expectedCache {
				t.Errorf("Expected cache '%s', got '%s'", tt.expectedCache, config.Cache)
			}

			if len(config.Commands) != tt.expectedCmdLen {
				t.Errorf("Expected %d commands, got %d", tt.expectedCmdLen, len(config.Commands))
			}

			if tt.expectedCmdLen > 0 {
				cmd := config.Commands[0]
				if cmd.Lang != tt.expectedLang {
					t.Errorf("Expected lang '%s', got '%s'", tt.expectedLang, cmd.Lang)
				}
				if cmd.GetExt() != tt.expectedExt {
					t.Errorf("Expected ext '%s', got '%s'", tt.expectedExt, cmd.GetExt())
				}
			}
		})
	}
}
