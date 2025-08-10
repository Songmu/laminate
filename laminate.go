package laminate

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const cmdName = "laminate"

// Run the laminate
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	ver := fs.Bool("version", false, "display version")
	lang := fs.String("lang", "", "code language (can also be set via CODEBLOCK_LANG env var)")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}

	// Get language from flag or environment
	// --lang flag takes precedence over CODEBLOCK_LANG environment variable
	var codeLang = os.Getenv("CODEBLOCK_LANG")
	if *lang != "" {
		codeLang = *lang
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if we have any commands configured
	if len(config.Commands) == 0 {
		return fmt.Errorf("no commands configured. Please create a config file at %s", getConfigPath())
	}

	// Read input from stdin
	var inputBuffer bytes.Buffer
	if _, err := io.Copy(&inputBuffer, os.Stdin); err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	input := inputBuffer.String()

	if input == "" {
		return fmt.Errorf("no input provided")
	}

	// Execute with cache support
	if err := ExecuteWithCache(ctx, config, codeLang, input, outStream); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}
