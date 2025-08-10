package laminate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/k1LoW/exec"
)

// Executor handles command execution
type Executor struct {
	cmd    *Command
	lang   string
	input  string
	output string
}

// NewExecutor creates a new executor
func NewExecutor(cmd *Command, lang, input, output string) *Executor {
	return &Executor{
		cmd:    cmd,
		lang:   lang,
		input:  input,
		output: output,
	}
}

// Execute runs the command and returns the output
func (e *Executor) Execute(ctx context.Context) ([]byte, error) {
	vars := PrepareVariables(e.input, e.output, e.lang)

	if e.cmd.Run.IsArray() {
		return e.executeArray(ctx, e.cmd.Run.Array(), vars)
	}
	return e.executeString(ctx, e.cmd.Run.String(), vars)
}

// executeString executes a string command
func (e *Executor) executeString(ctx context.Context, cmdStr string, vars map[string]string) ([]byte, error) {
	expanded, err := ExpandTemplate(cmdStr, vars)
	if err != nil {
		return nil, err
	}

	shell := e.cmd.GetShell()

	// Create command with shell
	cmd := exec.CommandContext(ctx, shell, "-c", expanded)

	// Set up stdin if {{input}} is not in the command
	if !HasVariable(cmdStr, "input") {
		cmd.Stdin = strings.NewReader(e.input)
	}

	// Set up output handling
	var outputBuffer bytes.Buffer
	if !HasVariable(cmdStr, "output") {
		// Command will output to stdout
		cmd.Stdout = &outputBuffer
	} else {
		// Command will write to file, we need to read it after
		cmd.Stdout = os.Stdout // For debugging
	}

	var errBuffer bytes.Buffer
	cmd.Stderr = &errBuffer

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, errBuffer.String())
	}

	// Get the output
	if !HasVariable(cmdStr, "output") {
		// Output was written to stdout
		return outputBuffer.Bytes(), nil
	} else {
		// Output was written to file, read it
		return os.ReadFile(e.output)
	}
}

// executeArray executes an array command
func (e *Executor) executeArray(ctx context.Context, cmdArray []string, vars map[string]string) ([]byte, error) {
	expanded, err := ExpandTemplateArray(cmdArray, vars)
	if err != nil {
		return nil, err
	}

	if len(expanded) == 0 {
		return nil, fmt.Errorf("empty command array")
	}

	// Create command
	cmd := exec.CommandContext(ctx, expanded[0], expanded[1:]...)

	// Set up stdin if {{input}} is not in the command
	if !HasVariableInArray(cmdArray, "input") {
		cmd.Stdin = strings.NewReader(e.input)
	}

	// Set up output handling
	var outputBuffer bytes.Buffer
	if !HasVariableInArray(cmdArray, "output") {
		// Command will output to stdout
		cmd.Stdout = &outputBuffer
	} else {
		// Command will write to file
		cmd.Stdout = os.Stdout // For debugging
	}

	var errBuffer bytes.Buffer
	cmd.Stderr = &errBuffer

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, errBuffer.String())
	}

	// Get the output
	if !HasVariableInArray(cmdArray, "output") {
		// Output was written to stdout
		return outputBuffer.Bytes(), nil
	} else {
		// Output was written to file, read it
		return os.ReadFile(e.output)
	}
}

// createTempFile creates a temporary file for output
func createTempFile(ext string) (string, error) {
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("laminate-*%s", ext))
	if err != nil {
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

// ExecuteWithCache executes a command with caching support
func ExecuteWithCache(ctx context.Context, config *Config, lang, input string, output io.Writer) error {
	// Find matching command
	cmd, err := FindMatchingCommand(config.Commands, lang)
	if err != nil {
		return err
	}

	ext := cmd.GetExt()

	// Parse cache duration
	duration, err := config.ParseDuration()
	if err != nil {
		return fmt.Errorf("failed to parse cache duration: %w", err)
	}

	// Create cache instance
	cache := NewCache(duration)

	// Try to get from cache
	if data, found := cache.Get(lang, input, ext); found {
		return copyToWriter(output, data)
	}

	// Create temporary output file
	outputPath, err := createTempFile("." + ext)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(outputPath)

	// Execute command
	executor := NewExecutor(cmd, lang, input, outputPath)
	data, err := executor.Execute(ctx)
	if err != nil {
		return err
	}

	// Store in cache
	if cacheErr := cache.Set(lang, input, ext, data); cacheErr != nil {
		// Log cache error but don't fail the operation
		fmt.Fprintf(os.Stderr, "Warning: failed to cache result: %v\n", cacheErr)
	}

	// Write to output
	return copyToWriter(output, data)
}
