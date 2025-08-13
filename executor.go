package laminate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(e.input)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	// If the result is written to a temporary file, read it from that file.
	if b, err := os.ReadFile(e.output); err == nil {
		fmt.Fprint(os.Stderr, buf.String())
		return b, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read output file: %w\nstdout: %s", err, buf.String())
	}
	// If it does not exist, read the result from stdout.
	return buf.Bytes(), nil
}

// ExecuteWithCache executes a command with caching support
func ExecuteWithCache(ctx context.Context, config *Config, lang, input string, output io.Writer) error {
	cmd, err := FindMatchingCommand(config.Commands, lang)
	if err != nil {
		return err
	}
	ext := cmd.GetExt()

	cache := NewCache(config.Cache)
	if data, found := cache.Get(lang, input, ext); found {
		_, err := output.Write(data)
		return err
	}

	tempDir, err := os.MkdirTemp("", "laminate-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	executor := &Executor{
		cmd:    cmd,
		lang:   lang,
		input:  input,
		output: filepath.Join(tempDir, "output."+ext),
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
