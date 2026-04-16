package tui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/list"
)

// normalizeForSearch lowercases s and keeps only letters and digits.
// It lets users search "entri-prod-main-service" as "entriprodmain".
func normalizeForSearch(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// normalizeWithMap returns the normalized form of s and a slice mapping
// each index in the normalized string back to its byte offset in s.
// Used to recover original positions for match highlighting.
func normalizeWithMap(s string) (string, []int) {
	var b strings.Builder
	b.Grow(len(s))
	idx := make([]int, 0, len(s))
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			idx = append(idx, i)
		}
	}
	return b.String(), idx
}

// matchesNormalized reports whether query matches haystack after stripping
// non-alphanumerics from both and lowercasing. Empty query matches everything.
func matchesNormalized(haystack, query string) bool {
	q := normalizeForSearch(query)
	if q == "" {
		return true
	}
	return strings.Contains(normalizeForSearch(haystack), q)
}

// normalizedSubstringFilter is a bubbles/list FilterFunc that matches on the
// alphanumeric-normalized form of each target. MatchedIndexes are mapped back
// to the original target bytes so the list delegate highlights correctly.
func normalizedSubstringFilter(term string, targets []string) []list.Rank {
	nTerm := normalizeForSearch(term)
	if nTerm == "" {
		ranks := make([]list.Rank, len(targets))
		for i := range targets {
			ranks[i] = list.Rank{Index: i}
		}
		return ranks
	}
	var ranks []list.Rank
	for i, t := range targets {
		nt, idxMap := normalizeWithMap(t)
		pos := strings.Index(nt, nTerm)
		if pos < 0 {
			continue
		}
		matched := make([]int, 0, len(nTerm))
		for k := 0; k < len(nTerm); k++ {
			matched = append(matched, idxMap[pos+k])
		}
		ranks = append(ranks, list.Rank{Index: i, MatchedIndexes: matched})
	}
	return ranks
}
