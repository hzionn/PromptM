package prompt

import (
	"path/filepath"
	"testing"
)

func TestLoadFromDirs(t *testing.T) {
	opts := Options{
		Extensions:     []string{".md", ".txt"},
		IgnorePatterns: []string{".DS_Store"},
		MaxFileSize:    128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")

	prompts, err := LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}

	found := make(map[string]Prompt)
	for _, p := range prompts {
		found[p.Name] = p
	}

	prompt, ok := found["product-brief"]
	if !ok {
		t.Fatalf("expected prompt product-brief to be discovered")
	}

	title, ok := prompt.FrontMatter["title"].(string)
	if !ok || title != "Product Brief" {
		t.Errorf("expected title %q, got %#v", "Product Brief", prompt.FrontMatter["title"])
	}

	summary, ok := prompt.FrontMatter["summary"].(string)
	if !ok || summary == "" {
		t.Errorf("expected summary to be populated, got %#v", prompt.FrontMatter["summary"])
	}

	metadata, ok := prompt.FrontMatter["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata map in front matter, got %#v", prompt.FrontMatter["metadata"])
	}
	if metadata["audience"] != "product" || metadata["stage"] != "discovery" {
		t.Errorf("unexpected metadata contents: %#v", metadata)
	}

	aliases, ok := prompt.FrontMatter["aliases"].([]any)
	if !ok || len(aliases) != 2 {
		t.Fatalf("expected aliases slice with 2 entries, got %#v", prompt.FrontMatter["aliases"])
	}

	if len(prompt.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(prompt.Tags))
	}

	if prompt.Content == "" {
		t.Fatal("expected prompt content to be populated")
	}

	if contentHasFrontMatter(prompt.Content) {
		t.Errorf("expected content to exclude front matter, got %q", prompt.Content)
	}
}

func TestLoadFromDirsRespectsExtensions(t *testing.T) {
	opts := Options{
		Extensions:     []string{".md"},
		IgnorePatterns: nil,
		MaxFileSize:    128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")

	prompts, err := LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	for _, p := range prompts {
		if filepath.Ext(p.Path) != ".md" {
			t.Errorf("unexpected extension %s in result", filepath.Ext(p.Path))
		}
	}
}

func contentHasFrontMatter(content string) bool {
	return len(content) > 0 && content[0] == '-'
}
