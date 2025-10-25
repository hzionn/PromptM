# PromptM - Prompt Manager CLI

> This project is still in development. Installation and settings are not convenient yet.

A fast, minimal CLI tool for managing and accessing your AI prompt library. PromptM lets you organize, search, and reuse prompts efficiently with fuzzy search and interactive selection.

## Motivation

pass

## Features

- 🔍 **Fuzzy Search** - Quickly find prompts with intelligent fuzzy matching
- 🎯 **Interactive Selection** - Beautiful TUI for browsing and selecting prompts
- 📋 **Multiple Commands** - Flexible CLI with `pick`, `search`, `ls`, `cat`, and `mesh` commands
- ⚙️ **Configurable** - Customize file extensions, directories, and search limits via `settings.toml`
- 📁 **Multi-Directory Support** - Load prompts from multiple directories
- 📋 **Clipboard Integration** - Copy selected prompts directly to clipboard
- 🚀 **Fast & Lightweight** - Single binary, no dependencies to install

## Installation

### From Source

Make sure you have [Go 1.24+](https://golang.org/doc/install) installed.

```bash
git clone https://github.com/hzionn/PromptM.git
cd PromptM
go build ./cmd/pm
```

This will produce a `pm` binary in the current directory.

Then point the prompt directory to your desired directory in `settings.toml`. Multiple directories are supported.

### Quick Start

```bash
# Make it globally accessible (optional)
sudo mv pm /usr/local/bin/pm

# Verify installation
pm --help
```

## Usage

### Basic Commands

#### Pick (Default)

Launch the interactive prompt picker:

```bash
pm
```

Or pick a prompt by query without interaction:

```bash
pm "code review"
pm --query "code review"
```

#### Search

Find prompts matching a query and display matches:

```bash
pm search "code review"
pm search --limit 5 "code"
```

Use `--interactive` flag to launch the picker after search:

```bash
pm search --interactive "code"
```

#### List

Show all available prompts:

```bash
pm ls
```

#### Cat

Display a specific prompt by name:

```bash
pm cat "code review"
```

#### Mesh

Combine multiple prompts together:

```bash
pm mesh "prompt1" "prompt2"
```

You can also pipe additional content:

```bash
pm mesh "system-prompt" "context-prompt" < user-input.txt
```

### Global Flags

- `--dir <paths>` - Override default prompt directories (comma-separated)
- `--query <query>` - Provide a query for non-interactive selection
- `--copy` - Copy the chosen prompt to clipboard
- `--interactive` - Force interactive selection mode

### Examples

```bash
# Pick a prompt interactively
pm

# Search for prompts about "testing"
pm search testing

# Get a specific prompt by name
pm cat "code-review"

# Search with a limit and use interactive picker
pm search --limit 10 --interactive "review"

# Copy a prompt to clipboard
pm --query "code review" --copy

# Use prompts from custom directories
pm --dir "/path/to/prompts,~/my-prompts" search "api"

# Combine multiple prompts
pm mesh "system-prompt" "user-prompt" | pbcopy
```

## Configuration

PromptM reads configuration from `config/settings.toml`. Create or modify this file to customize behavior:

```toml
# Default directories where prompts are stored
default_dir = ["./prompts"]

# Cache directory for temporary data
cache_dir = "./.pm-cache"

# File system settings
[file_system]
# File extensions to look for when scanning directories
extensions = [".md", ".txt"]

# Patterns to ignore when scanning
ignore_patterns = [".DS_Store"]

# Maximum file size to load (in KB)
max_file_size_kb = 128

# Fuzzy search configuration
[fuzzy_search]
# Maximum number of search results to return
max_results = 20

# UI configuration
[ui]
# Maximum length to truncate prompt display
truncate_length = 120
```

### Configuration Options

| Option                         | Type         | Description                                      |
| ------------------------------ | ------------ | ------------------------------------------------ |
| `default_dir`                  | Array/String | Directories to scan for prompts                  |
| `file_system.extensions`       | Array        | File extensions to include (e.g., `.md`, `.txt`) |
| `file_system.ignore_patterns`  | Array        | Glob patterns to exclude                         |
| `file_system.max_file_size_kb` | Number       | Maximum file size to load                        |
| `fuzzy_search.max_results`     | Number       | Max search results returned                      |
| `ui.truncate_length`           | Number       | Display truncation length                        |

## Project Structure

```
promptm/
├── cmd/pm/
│   └── main.go              # CLI entrypoint
├── internal/
│   ├── clipboard/           # Clipboard operations
│   ├── config/              # Configuration loading
│   ├── prompt/              # Prompt loading and management
│   ├── search/              # Fuzzy search implementation
│   └── ui/                  # Interactive TUI
├── config/
│   └── settings.toml        # Default configuration
├── prompts/                 # Example prompts
├── testdata/                # Test fixtures
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Development

### Building from Source

```bash
# Build the binary
go build ./cmd/pm

# Run tests
go test ./...

# Format code
find . -name '*.go' | xargs gofmt -w

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -run TestName ./...

# Format code
find . -name '*.go' | xargs gofmt -w
```

## Prompt Library Structure

Organize your prompts in directories with appropriate file extensions:

```
prompts/
├── code-review.md
├── brainstorm.txt
├── product-brief.md
└── stock_researcher.md
```

Prompt files can be in Markdown (`.md`) or text (`.txt`) format. The filename (without extension) becomes the prompt's name for selection.

### Example Prompt File

```markdown
# Code Review Checklist

- Verify naming conventions
- Ensure tests cover critical paths
- Confirm error handling
- Check for edge cases
```

## Dependencies

PromptM uses minimal dependencies:

- `github.com/charmbracelet/bubbletea` - Interactive TUI framework
- `github.com/lithammer/fuzzysearch` - Fuzzy search algorithm
- `github.com/pelletier/go-toml` - TOML configuration parsing
- `golang.org/x/term` - Terminal utilities

## Potential Roadmap

- [ ] Prompt tags and metadata
- [ ] Better UIUX
- [ ] Expose to package managers
- [ ] Color schemes and themes
- [ ] Batch operations on multiple prompts
- [ ] Plugin system for custom search strategies

## Support

Found a bug or have a feature request? Please open an [issue](https://github.com/hzionn/PromptM/issues).
