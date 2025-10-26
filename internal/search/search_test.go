package search

import (
	"path/filepath"
	"testing"

	"github.com/hzionn/prompt-manager-cli/internal/prompt"
)

func TestSearchMatchesByName(t *testing.T) {
	opts := prompt.Options{
		Extensions:  []string{".md", ".txt"},
		MaxFileSize: 128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")
	prompts, err := prompt.LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	results := Search(prompts, "product", Options{MaxResults: 5})
	if len(results) == 0 {
		t.Fatalf("expected results for query 'product'")
	}

	if results[0].Name != "product-brief" {
		t.Fatalf("expected top result to be product-brief, got %s", results[0].Name)
	}
}

func TestSearchEmptyQueryReturnsSorted(t *testing.T) {
	opts := prompt.Options{
		Extensions:  []string{".md", ".txt"},
		MaxFileSize: 128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")
	prompts, err := prompt.LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	results := Search(prompts, "", Options{})
	if len(results) != len(prompts) {
		t.Fatalf("expected %d results, got %d", len(prompts), len(results))
	}

	for i := 1; i < len(results); i++ {
		if results[i-1].Name > results[i].Name {
			t.Fatalf("expected results sorted by name, got %s before %s", results[i-1].Name, results[i].Name)
		}
	}
}

func TestSearchHandlesMisspellings(t *testing.T) {
	opts := prompt.Options{
		Extensions:  []string{".md", ".txt"},
		MaxFileSize: 128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")
	prompts, err := prompt.LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	results := Search(prompts, "prdct brf", Options{MaxResults: 3})
	if len(results) == 0 {
		t.Fatalf("expected fuzzy results for typo query")
	}

	if results[0].Name != "product-brief" {
		t.Fatalf("expected top fuzzy match to be product-brief, got %s", results[0].Name)
	}
}

func TestSearchUsesFrontMatterFields(t *testing.T) {
	opts := prompt.Options{
		Extensions:  []string{".md", ".txt"},
		MaxFileSize: 128 * 1024,
	}

	dir := filepath.FromSlash("../../testdata/prompts")
	prompts, err := prompt.LoadFromDirs([]string{dir}, opts)
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	results := Search(prompts, "discovery", Options{MaxResults: 5})
	if len(results) == 0 {
		t.Fatalf("expected results when querying front matter content")
	}

	if results[0].Name != "product-brief" {
		t.Fatalf("expected result to surface front matter match, got %s", results[0].Name)
	}
}
