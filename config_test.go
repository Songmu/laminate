package laminate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

func TestConfig_ParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		cache    string
		expected time.Duration
		hasError bool
	}{
		{"empty", "", 0, false},
		{"1h", "1h", time.Hour, false},
		{"30m", "30m", 30 * time.Minute, false},
		{"15s", "15s", 15 * time.Second, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Cache: tt.cache}
			duration, err := config.ParseDuration()

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if duration != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, duration)
				}
			}
		})
	}
}

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

func setupEnvVar(key, value string) func() {
	oldValue := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if oldValue != "" {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
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
			cleanup := setupEnvVar(tt.envVar, tt.envValue)
			defer cleanup()

			result := tt.getFunc()
			if result != tt.envValue {
				t.Errorf("Expected %s, got %s", tt.envValue, result)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    bool
		configContent  string
		expectedCache  string
		expectedCmdLen int
		expectedLang   string
		expectedExt    string
		expectError    bool
	}{
		{
			name:           "file_not_exists",
			setupConfig:    false,
			expectedCmdLen: 0,
			expectError:    false,
		},
		{
			name:        "valid_file",
			setupConfig: true,
			configContent: `cache: 1h
commands:
- lang: test
  run: echo test
  ext: png
`,
			expectedCache:  "1h",
			expectedCmdLen: 1,
			expectedLang:   "test",
			expectedExt:    "png",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()

			if tt.setupConfig {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")

				err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}

				cleanup = setupEnvVar("LAMINATE_CONFIG_PATH", configFile)
			} else {
				cleanup = setupEnvVar("LAMINATE_CONFIG_PATH", "/non/existent/config.yaml")
			}
			defer cleanup()

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

			if tt.expectedCache != "" && config.Cache != tt.expectedCache {
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
