laminate
========

[![Test Status](https://github.com/Songmu/laminate/actions/workflows/test.yaml/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/laminate/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/laminate)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/laminate)][PkgGoDev]

[actions]: https://github.com/Songmu/laminate/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/laminate
[license]: https://github.com/Songmu/laminate/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/laminate

A command-line bridge tool that orchestrates external image generation commands to convert text/code strings to images.

> [!IMPORTANT]
> `laminate` itself does not generate images. Instead, it acts as a **bridge** that routes input text to appropriate external tools (like `qrencode`, `silicon`, `mmdc`, `convert`, etc.) based on configurable patterns and manages the execution flow.

## How It Works

1. **Input**: Reads text from stdin and language specification via `--lang` flag or `CODEBLOCK_LANG` environment variable
2. **Routing**: Matches the language against configured patterns to select the appropriate external command
3. **Execution**: Runs the selected external command with proper input/output handling
4. **Output**: Returns the generated image data to stdout
5. **Caching**: Optionally caches results to avoid re-executing expensive commands

## Synopsis

```bash
# Generate QR code from text
echo "https://github.com/Songmu/laminate" | laminate --lang qr > qr.png

# Convert code to syntax-highlighted image
cat main.go | laminate --lang go > code.png

# Use with environment variable
export CODEBLOCK_LANG=python
cat script.py | laminate > python_code.png

# Generate image from any text (fallback to wildcard pattern)
echo "Hello World" | laminate --lang unknown > text.png

# Integration with k1LoW/deck for slide generation
deck apply -c laminate deck.md  # deck sets CODEBLOCK_LANG automatically
```

> [!TIP]
> `laminate` works seamlessly with [k1LoW/deck](https://github.com/k1LoW/deck) for generating slides with embedded code images. Use `deck apply -c laminate deck.md` to automatically convert code blocks in your markdown slides to images.

## Prerequisites

> [!IMPORTANT]
> You need to install the actual image generation tools that you want to use. `laminate` will fail if the required external commands are not available in your PATH.

The following are just examples of popular tools. You can use any command-line tool that can generate images - the choice is entirely up to you and your specific needs.

```bash
# For QR codes
brew install qrencode          # macOS
apt-get install qrencode       # Ubuntu/Debian

# For code syntax highlighting
cargo install silicon

# For Mermaid diagrams
npm install -g @mermaid-js/mermaid-cli

# For text-to-image (ImageMagick)
brew install imagemagick       # macOS
apt-get install imagemagick    # Ubuntu/Debian
```

## Installation

<details>
<summary>Click to expand installation methods</summary>

```console
# Install via Homebrew (macOS)
% brew install songmu/tap/laminate

# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/laminate/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/laminate/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/laminate/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/laminate/cmd/laminate@latest
```

</details>

## Configuration

Create a configuration file at `~/.config/laminate/config.yaml` (or `$XDG_CONFIG_HOME/laminate/config.yaml`):

```yaml
cache: 1h
commands:
- lang: qr
  run: 'qrencode -o "{{output}}" -t png "{{input}}"'
  ext: png
- lang: mermaid
  run: 'mmdc -i - -o "{{output}}" --quiet'
  ext: png
- lang: '{go,rust,python,java,javascript,typescript}'
  run: 'silicon -l "{{lang}}" -o "{{output}}"'
  ext: png
- lang: '*'
  run: ['convert', '-background', 'white', '-fill', 'black', 'label:{{input}}', '{{output}}']
```

### Configuration Schema

- **`cache`**: Cache duration (e.g., `1h`, `30m`, `15s`). Omit to disable caching.
- **`commands`**: Array of command configurations.
  - **`lang`**: Language pattern (supports glob patterns and brace expansion)
  - **`run`**: Command to execute (string or array format)
  - **`ext`**: Output file extension (default: `png`)
  - **`shell`**: Shell to use for string commands (default: `bash` or `sh`)

### Template Variables

You can use these variables in your commands as needed. The presence or absence of `{{input}}` and `{{output}}` determines how laminate handles I/O with the external command.

- **`{{input}}`**: Input text from stdin
  - Present: Input passed as command-line argument
  - Absent: Input passed via stdin to the command
- **`{{output}}`**: Output file path with extension from `ext` field (default: `png`)
  - Present: Command writes to this file, laminate reads it
  - Absent: Command writes to stdout, laminate captures it
- **`{{lang}}`**: The language parameter specified by user

**I/O Behavior Examples:**

| Variables Used | Example Command | How it works |
|---------------|-----------------|--------------|
| Both | `qrencode -o "{{output}}" "{{input}}"` | Input as arg, output to file |
| Output only | `mmdc -i - -o "{{output}}"` | Input via stdin, output to file |
| Input only | `convert label:"{{input}}" png:-` | Input as arg, output to stdout |
| Neither | `some-converter` | Input via stdin, output to stdout |

### Language Matching

Commands are matched against the specified language in **first-match-wins** order from top to bottom in the configuration file. The matching process:

1. **Sequential matching**: Each command's `lang` pattern is tested in the order they appear in the config
2. **First match wins**: The first command whose `lang` pattern matches the specified language is used
3. **Pattern types**: Supports exact matches, glob patterns, and brace expansion
   - Exact: `go`, `python`, `rust`
   - Brace expansion: `{go,rust,python}`, `{js,ts}`
   - Glob patterns: `py*`, `*script`, `*`
4. **Fallback**: Typically a wildcard pattern `*` is placed last to catch unmatched languages

**Example matching order:**
```yaml
commands:
  - lang: go            # 1st: Matches "go" exactly
  - lang: '{py,python}' # 2nd: Matches "py" or "python"
  - lang: 'js*'         # 3rd: Matches "js", "json", "jsx", etc.
  - lang: '*'           # 4th: Matches any remaining language
```

For language `python`: matches the 2nd command (`{py,python}`) and stops there.

> [!TIP]
> Put more specific patterns at the top and general patterns (like `*`) at the bottom to ensure proper matching priority.

## Environment Variables

- `CODEBLOCK_LANG`: Language specification via environment variable (automatically set by [k1LoW/deck](https://github.com/k1LoW/deck))

## Cache Management

Cache files are stored in `~/.cache/laminate/cache/` and keyed by input content + language + format.

```yaml
# Set cache duration
cache: 2h

# Disable caching (omit cache field)
# cache: 0s
```

## Usage Examples

### Template Variable Behaviors

#### Commands with `{{input}}` and `{{output}}`
```yaml
# Input passed as argument, output to file
- lang: qr
  run: 'qrencode -o "{{output}}" -t png "{{input}}"'
  ext: png
```
```bash
echo "https://example.com" | laminate --lang qr > qr.png
# Executes: qrencode -o "/tmp/laminate123.png" -t png "https://example.com"
```

#### Commands with `{{output}}` only (stdin input)
```yaml
# Input via stdin, output to file
- lang: mermaid
  run: 'mmdc -i - -o "{{output}}" --quiet'
  ext: png
```
```bash
echo "graph TD; A-->B" | laminate --lang mermaid > diagram.png
# Executes: mmdc -i - -o "/tmp/laminate456.png" --quiet
# (with "graph TD; A-->B" passed via stdin)
```

#### Commands without `{{output}}` (stdout output)
```yaml
# Input as argument, output via stdout
- lang: text
  run: 'convert -background white -fill black label:"{{input}}" png:-'
```
```bash
echo "Hello World" | laminate --lang text > text.png
# Executes: convert -background white -fill black label:"Hello World" png:-
# (image data read from command's stdout)
```

#### Commands without both variables (stdin to stdout)
```yaml
# Input via stdin, output via stdout
- lang: simple
  run: 'some-image-converter'
```
```bash
echo "input text" | laminate --lang simple > output.png
# Executes: some-image-converter
# (with "input text" passed via stdin, image read from stdout)
```

### Real-world Examples

#### QR Code Generation
```bash
echo "https://example.com" | laminate --lang qr > qr.png
```

#### Code Syntax Highlighting
```bash
# Using --lang flag (highest priority)
cat main.go | laminate --lang go > code.png

# Using environment variable
export CODEBLOCK_LANG=python
cat script.py | laminate > highlighted.png

# Empty language (uses first matching pattern)
cat README.md | laminate --lang "" > readme.png
```

> [!NOTE]
> Priority: `--lang` flag > `CODEBLOCK_LANG` environment variable > empty string

#### Mermaid Diagrams
```bash
cat << EOF | laminate --lang mermaid > diagram.png
graph TD
    A[Start] --> B[Process]
    B --> C[End]
EOF
```


## Author

[Songmu](https://github.com/Songmu)
