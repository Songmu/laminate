package laminate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
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

// Execute runs the command and returns the output
func (e *Executor) Execute(ctx context.Context) ([]byte, error) {
	argv, err := e.getArgv()
	if err != nil {
		return nil, fmt.Errorf("failed to get command arguments: %w", err)
	}
	return e.exceute(ctx, argv)
}

func (e *Executor) getArgv() ([]string, error) {
	vars := map[string]string{
		"input":  e.input,
		"output": e.output,
		"lang":   e.lang,
	}
	if e.cmd.Run.IsArray() {
		templates := e.cmd.Run.Array()
		var result = make([]string, len(templates))
		for i, template := range templates {
			expanded, err := ExpandTemplate(template, vars)
			if err != nil {
				return nil, err
			}
			result[i] = expanded
		}
		return result, nil
	}
	expanded, err := ExpandTemplate(e.cmd.Run.String(), vars)
	if err != nil {
		return nil, err
	}
	shell := e.cmd.GetShell()
	return []string{shell, "-c", expanded}, nil
}

func (e *Executor) exceute(ctx context.Context, argv []string) ([]byte, error) {
	var rawCmd = e.cmd.Run.Array()
	if !e.cmd.Run.IsArray() {
		rawCmd = []string{e.cmd.Run.String()}
	}

	hasInput := slices.ContainsFunc(rawCmd, func(arg string) bool {
		return HasVariable(arg, "input")
	})
	hasOutput := slices.ContainsFunc(rawCmd, func(arg string) bool {
		return HasVariable(arg, "output")
	})

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stderr = os.Stderr

	// Set up stdin if {{input}} is not in the command
	if !hasInput {
		cmd.Stdin = strings.NewReader(e.input)
	}

	// to prevent unnecessary output contamination, redirect stdout to stderr by deafult
	cmd.Stdout = os.Stderr

	// Set up output handling
	var outputBuffer bytes.Buffer
	if !hasOutput {
		// Command will output to stdout and we capture it
		cmd.Stdout = &outputBuffer
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	if !hasOutput {
		// Output was written to stdout
		return outputBuffer.Bytes(), nil
	}
	// Output was written to file, read it
	return os.ReadFile(e.output)
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
	cmd, err := FindMatchingCommand(config.Commands, lang)
	if err != nil {
		return err
	}
	ext := cmd.GetExt()

	duration, err := config.ParseDuration()
	if err != nil {
		return fmt.Errorf("failed to parse cache duration: %w", err)
	}

	cache := NewCache(duration)
	if data, found := cache.Get(lang, input, ext); found {
		_, err := output.Write(data)
		return err
	}

	outputPath, err := createTempFile("." + ext)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(outputPath)

	executor := &Executor{
		cmd:    cmd,
		lang:   lang,
		input:  input,
		output: outputPath,
	}
	data, err := executor.Execute(ctx)
	if err != nil {
		return err
	}

	if cacheErr := cache.Set(lang, input, ext, data); cacheErr != nil {
		// Log cache error but don't fail the operation
		fmt.Fprintf(os.Stderr, "Warning: failed to cache result: %v\n", cacheErr)
	}
	_, err = output.Write(data)
	return err
}
