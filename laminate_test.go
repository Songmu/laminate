package laminate_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Songmu/laminate"
)

func setupTestEnv(t *testing.T) (configPath, cachePath string) {
	t.Helper()

	// Create temporary directories
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "config.yaml")
	cachePath = filepath.Join(tmpDir, "cache")

	t.Setenv("LAMINATE_CONFIG_PATH", configPath)
	t.Setenv("LAMINATE_CACHE_PATH", cachePath)

	return configPath, cachePath
}

func createTestConfigFromFile(t *testing.T, configPath, configType string) {
	t.Helper()

	var sourceConfig string
	switch configType {
	case "default":
		sourceConfig = "testdata/test_config.yaml"
	case "no_cache":
		sourceConfig = "testdata/test_config_no_cache.yaml"
	case "asterisk_first":
		sourceConfig = "testdata/test_config_asterisk_first.yaml"
	default:
		t.Fatalf("Unknown config type: %s", configType)
	}

	configContent, err := os.ReadFile(sourceConfig)
	if err != nil {
		t.Fatalf("Failed to read test config template: %v", err)
	}

	err = os.WriteFile(configPath, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
}

func setupStdinWithInput(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	if input != "" {
		go func() {
			w.Write([]byte(input))
			w.Close()
		}()
	} else {
		w.Close()
	}

	return func() { os.Stdin = oldStdin }
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

func assertImageFormat(t *testing.T, output []byte, format string) {
	t.Helper()
	if len(output) == 0 {
		t.Error("Expected output, got empty")
		return
	}

	switch format {
	case "png":
		if len(output) < 8 || output[0] != 0x89 || output[1] != 0x50 || output[2] != 0x4E || output[3] != 0x47 {
			t.Errorf("Expected PNG signature, got: %v", output[:min(8, len(output))])
		}
	case "jpg", "jpeg":
		if len(output) < 4 || output[0] != 0xFF || output[1] != 0xD8 || output[2] != 0xFF {
			t.Errorf("Expected JPEG signature, got: %v", output[:min(4, len(output))])
		}
	default:
		t.Errorf("Unknown image format: %s", format)
	}
}

func TestRun_Version(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	err := laminate.Run(context.Background(), []string{"--version"}, &outBuf, &errBuf)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "laminate") {
		t.Errorf("Expected version output to contain 'laminate', got: %s", output)
	}
}

func TestRun_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    bool
		args           []string
		input          string
		expectedErrMsg string
	}{
		{
			name:           "no_config_file",
			setupConfig:    false,
			args:           []string{"--lang", "test"},
			input:          "test input",
			expectedErrMsg: "failed to read config file",
		},
		{
			name:           "no_input_provided",
			setupConfig:    true,
			args:           []string{"--lang", "text"},
			input:          "",
			expectedErrMsg: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, _ := setupTestEnv(t)

			if tt.setupConfig {
				createTestConfigFromFile(t, configPath, "default")
			}

			var outBuf, errBuf bytes.Buffer

			// Set stdin
			cleanupStdin := setupStdinWithInput(tt.input)
			defer cleanupStdin()

			err := laminate.Run(context.Background(), tt.args, &outBuf, &errBuf)
			if err == nil {
				t.Errorf("Expected error, got nil: %s", tt.name)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestRun_ImageGeneration(t *testing.T) {
	tests := []struct {
		name       string
		configType string
		args       []string
		envLang    string
		input      string
		format     string
	}{
		{
			name:       "go_language_png",
			configType: "default",
			args:       []string{"--lang", "go"},
			input:      "package main\n\nfunc main() {}",
			format:     "png",
		},
		{
			name:       "python_with_env_var",
			configType: "default",
			args:       []string{},
			envLang:    "python",
			input:      "print('hello')",
			format:     "png",
		},
		{
			name:       "rust_language_jpg",
			configType: "default",
			args:       []string{"--lang", "rust"},
			input:      "fn main() { println!(\"Hello, Rust!\"); }",
			format:     "jpg",
		},
		{
			name:       "brace_expansion_default_extension",
			configType: "default",
			args:       []string{"--lang", "java"},
			input:      "public class Hello { }",
			format:     "png",
		},
		{
			name:       "empty_lang_string",
			configType: "default",
			args:       []string{"--lang", ""},
			input:      "content with empty lang",
			format:     "png",
		},
		{
			name:       "wildcard_pattern",
			configType: "default",
			args:       []string{"--lang", "unknown"},
			input:      "unknown language content",
			format:     "png",
		},
		{
			name:       "pattern_matching_priority_asterisk_first",
			configType: "asterisk_first",
			args:       []string{"--lang", ""},
			input:      "pattern priority test content",
			format:     "jpg",
		},
		{
			name:       "lang_flag_precedence_over_env",
			configType: "default",
			args:       []string{"--lang", "go"},
			envLang:    "python", // Should be ignored since --lang is specified
			input:      "package main",
			format:     "png", // Go lang should produce PNG
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, _ := setupTestEnv(t)

			createTestConfigFromFile(t, configPath, tt.configType)

			// Set environment variable if specified
			if tt.envLang != "" {
				cleanupEnv := setupEnvVar("CODEBLOCK_LANG", tt.envLang)
				defer cleanupEnv()
			}

			var outBuf, errBuf bytes.Buffer

			// Set stdin
			cleanupStdin := setupStdinWithInput(tt.input)
			defer cleanupStdin()

			err := laminate.Run(context.Background(), tt.args, &outBuf, &errBuf)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Verify output format
			assertImageFormat(t, outBuf.Bytes(), tt.format)
		})
	}
}

func TestRun_CacheBehavior(t *testing.T) {
	tests := []struct {
		name        string
		configType  string
		lang        string
		expectCache bool
	}{
		{
			name:        "with_cache_enabled",
			configType:  "default",
			lang:        "text",
			expectCache: true,
		},
		{
			name:        "with_cache_disabled",
			configType:  "no_cache",
			lang:        "go",
			expectCache: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, _ := setupTestEnv(t)

			createTestConfigFromFile(t, configPath, tt.configType)

			input := "cache test content"

			// First run
			var outBuf1, errBuf1 bytes.Buffer
			cleanupStdin1 := setupStdinWithInput(input)

			start1 := time.Now()
			err := laminate.Run(context.Background(), []string{"--lang", tt.lang}, &outBuf1, &errBuf1)
			duration1 := time.Since(start1)
			cleanupStdin1()

			if err != nil {
				t.Errorf("First run failed: %v", err)
			}

			// Second run
			var outBuf2, errBuf2 bytes.Buffer
			cleanupStdin2 := setupStdinWithInput(input)

			start2 := time.Now()
			err = laminate.Run(context.Background(), []string{"--lang", tt.lang}, &outBuf2, &errBuf2)
			duration2 := time.Since(start2)
			cleanupStdin2()

			if err != nil {
				t.Errorf("Second run failed: %v", err)
			}

			// Compare outputs - should always be the same
			if !bytes.Equal(outBuf1.Bytes(), outBuf2.Bytes()) {
				t.Error("Both runs should generate same output")
			}

			if tt.expectCache {
				t.Logf("With cache - First run: %v, Second run: %v", duration1, duration2)
			} else {
				t.Logf("No cache - First run: %v, Second run: %v", duration1, duration2)
			}

			// Verify outputs are valid images
			assertImageFormat(t, outBuf1.Bytes(), "png")
			assertImageFormat(t, outBuf2.Bytes(), "png")
		})
	}
}
