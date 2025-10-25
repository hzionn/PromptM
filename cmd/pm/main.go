package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/hzionn/promptm/internal/clipboard"
	"github.com/hzionn/promptm/internal/config"
	"github.com/hzionn/promptm/internal/prompt"
	"github.com/hzionn/promptm/internal/search"
	"github.com/hzionn/promptm/internal/ui"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type appContext struct {
	settings   config.Settings
	promptOpts prompt.Options
	searchOpts search.Options
}

func newAppContext() appContext {
	settings := config.Load("config/settings.toml")
	maxBytes := int64(settings.FileSystem.MaxFileSizeKB) * 1024
	return appContext{
		settings: settings,
		promptOpts: prompt.Options{
			Extensions:     settings.FileSystem.Extensions,
			IgnorePatterns: settings.FileSystem.IgnorePatterns,
			MaxFileSize:    maxBytes,
		},
		searchOpts: search.Options{
			MaxResults: settings.FuzzySearch.MaxResults,
		},
	}
}

func run(args []string, in io.Reader, out io.Writer) error {
	ctx := newAppContext()
	if len(args) == 0 {
		return runPick(ctx, []string{}, in, out)
	}

	switch args[0] {
	case "pick":
		return runPick(ctx, args[1:], in, out)
	case "search":
		return runSearch(ctx, args[1:], in, out)
	case "ls":
		return runList(ctx, args[1:], out)
	case "cat":
		return runCat(ctx, args[1:], out)
	case "mesh":
		return runMesh(ctx, args[1:], in, out)
	case "--help", "-h", "help":
		printUsage(out)
		return nil
	default:
		if strings.HasPrefix(args[0], "-") {
			return runPick(ctx, args, in, out)
		}
		query := strings.Join(args, " ")
		return runPickWithQuery(ctx, query, "", false, out)
	}
}

func runPick(ctx appContext, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("pick", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var dirFlag string
	var query string
	var interactive bool
	var copyToClipboard bool

	fs.StringVar(&dirFlag, "dir", "", "Prompt directories (comma separated)")
	fs.StringVar(&query, "query", "", "Query to select a prompt non-interactively")
	fs.BoolVar(&interactive, "interactive", false, "Force interactive selection")
	fs.BoolVar(&copyToClipboard, "copy", false, "Copy the chosen prompt to the clipboard")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if query != "" && interactive {
		return errors.New("cannot use --query and --interactive together")
	}

	if query != "" {
		return runPickWithQuery(ctx, query, dirFlag, copyToClipboard, out)
	}

	if !interactive && fs.NArg() > 0 {
		// Allow positional query arguments.
		query = strings.Join(fs.Args(), " ")
		return runPickWithQuery(ctx, query, dirFlag, copyToClipboard, out)
	}

	return runPickInteractive(ctx, dirFlag, copyToClipboard, in, out)
}

func runPickWithQuery(ctx appContext, query, dirFlag string, copyToClipboard bool, out io.Writer) error {
	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	results := search.Search(prompts, query, ctx.searchOpts)
	if len(results) == 0 {
		return fmt.Errorf("no prompts found for query %q", query)
	}

	return outputPrompt(results[0].Content, copyToClipboard, out)
}

func runPickInteractive(ctx appContext, dirFlag string, copyToClipboard bool, in io.Reader, out io.Writer) error {
	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	if len(prompts) == 0 {
		return errors.New("no prompts available")
	}

	sorted := search.Search(prompts, "", search.Options{})
	// Use stderr for the interactive UI to keep stdout clean for the prompt output
	selected, err := ui.SelectPromptWithQuery(sorted, "", search.Options{}, in, os.Stderr)
	if err != nil {
		return err
	}

	return outputPrompt(selected.Content, copyToClipboard, out)
}

type fdReader interface {
	Fd() uintptr
}

type terminalAwareReader interface {
	IsTerminal() bool
}

func shouldReadFromInput(in io.Reader) bool {
	if in == nil {
		return false
	}

	if aware, ok := in.(terminalAwareReader); ok {
		return !aware.IsTerminal()
	}

	if fd, ok := in.(fdReader); ok {
		return !term.IsTerminal(int(fd.Fd()))
	}

	return true
}

func runSearch(ctx appContext, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var dirFlag string
	var limit int
	var interactive bool
	fs.StringVar(&dirFlag, "dir", "", "Prompt directories (comma separated)")
	fs.IntVar(&limit, "limit", ctx.searchOpts.MaxResults, "Maximum number of results")
	fs.BoolVar(&interactive, "interactive", false, "Launch interactive picker with the query")

	if err := fs.Parse(args); err != nil {
		return err
	}

	queryArgs := fs.Args()
	if len(queryArgs) == 0 {
		return errors.New("search requires a query argument")
	}
	query := strings.Join(queryArgs, " ")

	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	opts := ctx.searchOpts
	if limit > 0 {
		opts.MaxResults = limit
	}

	results := search.Search(prompts, query, opts)
	if len(results) == 0 {
		return fmt.Errorf("no prompts found for query %q", query)
	}

	if interactive {
		// Use stderr for the interactive UI to keep stdout clean for the prompt output
		selected, err := ui.SelectPromptWithQuery(prompts, query, opts, in, os.Stderr)
		if err != nil {
			return err
		}
		return writePrompt(out, selected.Content)
	}

	for _, p := range results {
		fmt.Fprintf(out, "%s\t%s\n", p.Name, p.Path)
	}
	return nil
}

func runList(ctx appContext, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var dirFlag string
	fs.StringVar(&dirFlag, "dir", "", "Prompt directories (comma separated)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	results := search.Search(prompts, "", search.Options{})
	for _, p := range results {
		fmt.Fprintln(out, p.Name)
	}
	return nil
}

func runCat(ctx appContext, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("cat", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var dirFlag string
	fs.StringVar(&dirFlag, "dir", "", "Prompt directories (comma separated)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		return errors.New("cat requires a prompt name")
	}
	name := strings.Join(names, " ")

	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	promptItem, ok := findPromptByName(prompts, name)
	if !ok {
		return fmt.Errorf("prompt %q not found", name)
	}

	return writePrompt(out, promptItem.Content)
}

func runMesh(ctx appContext, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("mesh", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var dirFlag string
	fs.StringVar(&dirFlag, "dir", "", "Prompt directories (comma separated)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		return errors.New("mesh requires at least one prompt name")
	}

	prompts, err := loadPrompts(ctx, dirFlag)
	if err != nil {
		return err
	}

	for _, name := range names {
		promptItem, ok := findPromptByName(prompts, name)
		if !ok {
			return fmt.Errorf("prompt %q not found", name)
		}
		if err := writePrompt(out, promptItem.Content); err != nil {
			return err
		}
		fmt.Fprintln(out)
	}

	if shouldReadFromInput(in) {
		if extra, err := io.ReadAll(in); err == nil && len(extra) > 0 {
			if err := writePrompt(out, string(extra)); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadPrompts(ctx appContext, dirFlag string) ([]prompt.Prompt, error) {
	dirs := ctx.settings.DefaultDirs
	if dirFlag != "" {
		dirs = splitAndTrim(dirFlag)
	}
	return prompt.LoadFromDirs(dirs, ctx.promptOpts)
}

func findPromptByName(prompts []prompt.Prompt, name string) (prompt.Prompt, bool) {
	for _, p := range prompts {
		if strings.EqualFold(p.Name, name) {
			return p, true
		}
	}
	return prompt.Prompt{}, false
}

func splitAndTrim(input string) []string {
	parts := strings.Split(input, ",")
	var out []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, `pm - prompt manager CLI

Usage:
  pm [--query <query>] [--dir <dir>]
  pm pick [--query <query>] [--interactive]
  pm search [--limit N] <query>
  pm ls
  pm cat <name>
  pm mesh <name> [<name>...]

Flags:
  --dir           Override prompt directories (comma separated)
  --query         Provide a query for prompt selection`)
}

func outputPrompt(content string, copyToClipboard bool, out io.Writer) error {
	cleaned := normalizeContent(content)
	if err := writePrompt(out, cleaned); err != nil {
		return err
	}
	if copyToClipboard {
		if err := clipboard.Copy(cleaned); err != nil {
			return fmt.Errorf("copy to clipboard: %w", err)
		}
	}
	return nil
}

func normalizeContent(content string) string {
	return strings.TrimRight(content, "\r\n")
}

func writePrompt(out io.Writer, content string) error {
	_, err := fmt.Fprintln(out, normalizeContent(content))
	return err
}
