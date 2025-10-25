package search

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/hzionn/promptm/internal/prompt"
)

var normalizerReplacer = strings.NewReplacer(
	"_", " ",
	"-", " ",
	"/", " ",
	".", " ",
	",", " ",
	"\n", " ",
	"\r", " ",
	"\t", " ",
)

// Options configure search behaviour.
type Options struct {
	MaxResults int
}

// Search applies fuzzy matching to find prompts that best align with the query.
func Search(prompts []prompt.Prompt, query string, opts Options) []prompt.Prompt {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		results := append([]prompt.Prompt(nil), prompts...)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
		if opts.MaxResults > 0 && len(results) > opts.MaxResults {
			return results[:opts.MaxResults]
		}
		return results
	}

	qNorm := normalize(trimmed)

	type scored struct {
		score  float64
		prompt prompt.Prompt
	}

	var matches []scored
	for _, p := range prompts {
		score := aggregateScore(p, trimmed, qNorm)
		if score <= 0 {
			continue
		}
		matches = append(matches, scored{score: score, prompt: p})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if almostEqual(matches[i].score, matches[j].score) {
			return matches[i].prompt.Name < matches[j].prompt.Name
		}
		return matches[i].score > matches[j].score
	})

	if opts.MaxResults > 0 && len(matches) > opts.MaxResults {
		matches = matches[:opts.MaxResults]
	}

	results := make([]prompt.Prompt, 0, len(matches))
	for _, m := range matches {
		results = append(results, m.prompt)
	}
	return results
}

func aggregateScore(p prompt.Prompt, rawQuery, normalizedQuery string) float64 {
	nameScore := fuzzyScore(normalizedQuery, normalize(p.Name))

	aliasScore := bestScore(valuesFromFront(p.FrontMatter, "aliases"), normalizedQuery)
	tagScore := bestScore(p.Tags, normalizedQuery)

	metaSkip := map[string]struct{}{
		"tags":    {},
		"aliases": {},
	}
	metaScore := bestScore(collectFrontMatterStrings(p.FrontMatter, metaSkip), normalizedQuery)

	contentScore := contentRelevance(p.Content, rawQuery, normalizedQuery)

	total := (nameScore * 5.0) +
		(aliasScore * 4.0) +
		(tagScore * 3.0) +
		(metaScore * 2.0) +
		contentScore

	if total < 0.25 {
		return 0
	}
	return total
}

func bestScore(values []string, query string) float64 {
	best := 0.0
	for _, value := range values {
		score := fuzzyScore(query, normalize(value))
		if score > best {
			best = score
		}
	}
	return best
}

func contentRelevance(content, rawQuery, normalizedQuery string) float64 {
	if content == "" || normalizedQuery == "" {
		return 0
	}

	contentNorm := normalize(snippetForSearch(content, 2048))
	if contentNorm == "" {
		return 0
	}

	if strings.Contains(contentNorm, normalizedQuery) {
		return 1.5
	}

	// Give a slight boost if the raw query appears in the original content (case insensitive).
	if strings.Contains(strings.ToLower(content), strings.ToLower(rawQuery)) {
		return 1.0
	}

	return fuzzyScore(normalizedQuery, contentNorm) * 0.75
}

func fuzzyScore(query, candidate string) float64 {
	if query == "" || candidate == "" {
		return 0
	}

	if strings.Contains(candidate, query) {
		return 1
	}

	distance := fuzzy.RankMatchNormalizedFold(query, candidate)
	if distance < 0 {
		return 0
	}

	score := 1.0 / (1.0 + float64(distance))
	if score > 1 {
		return 1
	}
	return score
}

func valuesFromFront(front map[string]any, key string) []string {
	if front == nil {
		return nil
	}
	value, ok := front[key]
	if !ok {
		return nil
	}
	return flattenValue(value)
}

func collectFrontMatterStrings(front map[string]any, skip map[string]struct{}) []string {
	if front == nil {
		return nil
	}

	var values []string
	for key, val := range front {
		if skip != nil {
			if _, ok := skip[strings.ToLower(key)]; ok {
				continue
			}
		}
		values = append(values, flattenValue(val)...)
	}
	return values
}

func flattenValue(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	case []string:
		return append([]string(nil), v...)
	case []any:
		var result []string
		for _, item := range v {
			result = append(result, flattenValue(item)...)
		}
		return result
	case map[string]any:
		var result []string
		for _, item := range v {
			result = append(result, flattenValue(item)...)
		}
		return result
	default:
		return []string{fmt.Sprint(v)}
	}
}

func snippetForSearch(content string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(content)
	if len(runes) <= maxRunes {
		return content
	}
	return string(runes[:maxRunes])
}

func normalize(value string) string {
	if value == "" {
		return ""
	}
	clean := normalizerReplacer.Replace(strings.ToLower(value))
	return strings.Join(strings.Fields(clean), " ")
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-6
}
