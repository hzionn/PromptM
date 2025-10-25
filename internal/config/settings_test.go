package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	settings := Load("non-existent.toml")

	if len(settings.DefaultDirs) == 0 {
		t.Fatal("expected default directories to be populated")
	}
	if settings.FileSystem.MaxFileSizeKB == 0 {
		t.Fatal("expected default file size to be populated")
	}
}

func TestLoadParsesStringAndSliceDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.toml")

	content := []byte(`
default_dir = "custom"

[file_system]
extensions = [".md"]
ignore_patterns = ["*.tmp"]
max_file_size_kb = 42

[fuzzy_search]
max_results = 5
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	settings := Load(path)

	if len(settings.DefaultDirs) != 1 || settings.DefaultDirs[0] != "custom" {
		t.Fatalf("expected default dir 'custom', got %v", settings.DefaultDirs)
	}

	if settings.FileSystem.MaxFileSizeKB != 42 {
		t.Fatalf("expected MaxFileSizeKB 42, got %d", settings.FileSystem.MaxFileSizeKB)
	}

	if settings.FuzzySearch.MaxResults != 5 {
		t.Fatalf("expected MaxResults 5, got %d", settings.FuzzySearch.MaxResults)
	}
}
