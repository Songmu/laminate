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

func TestGetConfigPath_EnvOverride(t *testing.T) {
	oldPath := os.Getenv("LAMINATE_CONFIG_PATH")
	defer func() {
		if oldPath != "" {
			os.Setenv("LAMINATE_CONFIG_PATH", oldPath)
		} else {
			os.Unsetenv("LAMINATE_CONFIG_PATH")
		}
	}()

	testPath := "/tmp/test-config.yaml"
	os.Setenv("LAMINATE_CONFIG_PATH", testPath)

	result := getConfigPath()
	if result != testPath {
		t.Errorf("Expected %s, got %s", testPath, result)
	}
}

func TestGetCachePath_EnvOverride(t *testing.T) {
	oldPath := os.Getenv("LAMINATE_CACHE_PATH")
	defer func() {
		if oldPath != "" {
			os.Setenv("LAMINATE_CACHE_PATH", oldPath)
		} else {
			os.Unsetenv("LAMINATE_CACHE_PATH")
		}
	}()

	testPath := "/tmp/test-cache"
	os.Setenv("LAMINATE_CACHE_PATH", testPath)

	result := getCachePath()
	if result != testPath {
		t.Errorf("Expected %s, got %s", testPath, result)
	}
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	oldPath := os.Getenv("LAMINATE_CONFIG_PATH")
	defer func() {
		if oldPath != "" {
			os.Setenv("LAMINATE_CONFIG_PATH", oldPath)
		} else {
			os.Unsetenv("LAMINATE_CONFIG_PATH")
		}
	}()

	// Set non-existent file
	os.Setenv("LAMINATE_CONFIG_PATH", "/non/existent/config.yaml")

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if config == nil {
		t.Error("Expected empty config, got nil")
	} else if len(config.Commands) != 0 {
		t.Error("Expected empty commands")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `cache: 1h
commands:
- lang: test
  run: echo test
  ext: png
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	oldPath := os.Getenv("LAMINATE_CONFIG_PATH")
	defer func() {
		if oldPath != "" {
			os.Setenv("LAMINATE_CONFIG_PATH", oldPath)
		} else {
			os.Unsetenv("LAMINATE_CONFIG_PATH")
		}
	}()

	os.Setenv("LAMINATE_CONFIG_PATH", configFile)

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if config.Cache != "1h" {
		t.Errorf("Expected cache '1h', got '%s'", config.Cache)
	}

	if len(config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(config.Commands))
	}

	cmd := config.Commands[0]
	if cmd.Lang != "test" {
		t.Errorf("Expected lang 'test', got '%s'", cmd.Lang)
	}

	if cmd.GetExt() != "png" {
		t.Errorf("Expected ext 'png', got '%s'", cmd.GetExt())
	}
}
