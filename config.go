package laminate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/pathologize"
)

// Config represents the configuration for laminate
type Config struct {
	Cache    time.Duration `yaml:"cache"`
	Commands []*Command    `yaml:"commands"`
}

// RunCommand represents a command that can be either a string or []string
type RunCommand struct {
	isArray bool
	str     string
	array   []string
}

// UnmarshalYAML implements yaml.Unmarshaler
func (r *RunCommand) UnmarshalYAML(unmarshal func(any) error) error {
	// Try to unmarshal as string first
	var str string
	if err := unmarshal(&str); err == nil {
		r.str = str
		r.isArray = false
		return nil
	}

	// Try to unmarshal as []string
	var array []string
	if err := unmarshal(&array); err == nil {
		r.array = array
		r.isArray = true
		return nil
	}

	return fmt.Errorf("run must be string or array of strings")
}

// IsArray returns true if the command is an array
func (r *RunCommand) IsArray() bool {
	return r.isArray
}

// String returns the command as string (only valid if IsArray() == false)
func (r *RunCommand) String() string {
	return r.str
}

// Array returns the command as []string (only valid if IsArray() == true)
func (r *RunCommand) Array() []string {
	return r.array
}

// Command represents a single command configuration
type Command struct {
	Lang  string     `yaml:"lang"`
	Run   RunCommand `yaml:"run"`
	Ext   string     `yaml:"ext"`
	Shell string     `yaml:"shell"`
}

// GetExt returns the file extension for the output
func (cmd *Command) GetExt() string {
	if cmd.Ext != "" {
		return pathologize.Clean(cmd.Ext)
	}
	return "png"
}

// GetShell returns the shell to use for command execution
func (cmd *Command) GetShell() string {
	if cmd.Shell != "" {
		return cmd.Shell
	}
	if path, err := exec.LookPath("bash"); err == nil {
		return path
	}
	// Fallback to sh
	if path, err := exec.LookPath("sh"); err == nil {
		return path
	}
	return "/bin/sh"
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	if configPath := os.Getenv("LAMINATE_CONFIG_PATH"); configPath != "" {
		return configPath
	}

	if runtime.GOOS == "windows" {
		configDir, _ := os.UserConfigDir()
		return filepath.Join(configDir, "laminate", "config.yaml")
	}

	// Use XDG_CONFIG_HOME for non-Windows
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "laminate", "config.yaml")
}

// getCachePath returns the path to the cache directory
func getCachePath() string {
	if cachePath := os.Getenv("LAMINATE_CACHE_PATH"); cachePath != "" {
		return cachePath
	}

	if runtime.GOOS == "windows" {
		cacheDir, _ := os.UserCacheDir()
		return filepath.Join(cacheDir, "laminate", "cache")
	}

	// Use XDG_CACHE_HOME for non-Windows
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "laminate", "cache")
}
