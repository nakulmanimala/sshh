package tui

import "testing"

func TestMatchesNormalized(t *testing.T) {
	const name = "entri-prod-main-service"
	cases := []struct {
		query string
		want  bool
	}{
		{"entriprodmain", true},
		{"prodmain", true},
		{"ENTRI-PROD", true}, // separators and case ignored on both sides
		{"mainentri", false},
		{"epms", false},
		{"", true},
	}
	for _, c := range cases {
		if got := matchesNormalized(name, c.query); got != c.want {
			t.Errorf("matchesNormalized(%q, %q) = %v, want %v", name, c.query, got, c.want)
		}
	}
}

func TestNormalizedSubstringFilterMapsIndexes(t *testing.T) {
	targets := []string{"entri-prod-main-service", "other-server"}
	ranks := normalizedSubstringFilter("prodmain", targets)
	if len(ranks) != 1 || ranks[0].Index != 0 {
		t.Fatalf("expected one match at index 0, got %+v", ranks)
	}
	// Matched indexes must point back into the original string and cover
	// the exact "prod" + "main" spans.
	got := ranks[0].MatchedIndexes
	want := []int{6, 7, 8, 9, 11, 12, 13, 14}
	if len(got) != len(want) {
		t.Fatalf("MatchedIndexes len=%d want %d: %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("MatchedIndexes[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
