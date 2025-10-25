package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hzionn/promptm/internal/clipboard"
	"github.com/hzionn/promptm/internal/config"
	"github.com/hzionn/promptm/internal/prompt"
	"github.com/hzionn/promptm/internal/search"
)

type terminalStub struct {
	readCalled bool
}

func (t *terminalStub) Read(p []byte) (int, error) {
	t.readCalled = true
	return 0, nil
}

func (t *terminalStub) IsTerminal() bool { return true }

func TestRunMeshSkipsReadingWhenInputIsTerminal(t *testing.T) {
	ctx := testAppContext()
	var out bytes.Buffer
	stub := &terminalStub{}

	if err := runMesh(ctx, []string{"code-review"}, stub, &out); err != nil {
		t.Fatalf("runMesh error = %v", err)
	}

	if stub.readCalled {
		t.Fatal("expected terminal input not to be read")
	}
}

func TestRunMeshReadsPipedInput(t *testing.T) {
	ctx := testAppContext()
	var out bytes.Buffer
	input := strings.NewReader("Additional context\n")

	if err := runMesh(ctx, []string{"code-review"}, input, &out); err != nil {
		t.Fatalf("runMesh error = %v", err)
	}

	if !strings.Contains(out.String(), "Additional context") {
		t.Fatalf("expected mesh output to include piped input, got %q", out.String())
	}
}

func TestRunPickWithQueryCopiesSelection(t *testing.T) {
	ctx := testAppContext()
	var copied string
	clipboard.SetProvider(clipboard.ProviderFunc(func(text string) error {
		copied = text
		return nil
	}))
	defer clipboard.SetProvider(nil)

	var out bytes.Buffer
	if err := runPickWithQuery(ctx, "code-review", "", true, &out); err != nil {
		t.Fatalf("runPickWithQuery error = %v", err)
	}

	if copied == "" || !strings.Contains(copied, "Code Review") {
		t.Fatalf("expected clipboard to receive prompt content, got %q", copied)
	}
}

func TestRunPickInteractiveCopiesSelection(t *testing.T) {
	ctx := testAppContext()
	var copied string
	clipboard.SetProvider(clipboard.ProviderFunc(func(text string) error {
		copied = text
		return nil
	}))
	defer clipboard.SetProvider(nil)

	input := strings.NewReader("1\n")
	var out bytes.Buffer

	if err := runPickInteractive(ctx, "", true, input, &out); err != nil {
		t.Fatalf("runPickInteractive error = %v", err)
	}

	if copied == "" {
		t.Fatal("expected clipboard to be populated for interactive pick")
	}
}

func testAppContext() appContext {
	dir := filepath.Join("..", "..", "testdata", "prompts")
	settings := config.Settings{
		DefaultDirs: []string{dir},
		FileSystem: config.FileSystemSettings{
			Extensions:     []string{".md", ".txt"},
			IgnorePatterns: nil,
			MaxFileSizeKB:  1024,
		},
		FuzzySearch: config.FuzzySearchSettings{MaxResults: 20},
	}

	return appContext{
		settings: settings,
		promptOpts: prompt.Options{
			Extensions:     settings.FileSystem.Extensions,
			IgnorePatterns: settings.FileSystem.IgnorePatterns,
			MaxFileSize:    int64(settings.FileSystem.MaxFileSizeKB) * 1024,
		},
		searchOpts: search.Options{MaxResults: settings.FuzzySearch.MaxResults},
	}
}
