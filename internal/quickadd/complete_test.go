package quickadd

import (
	"slices"
	"testing"
	"unicode/utf8"
)

func TestTagSuggestionsRespectInputContext(t *testing.T) {
	for _, tc := range []struct {
		line, project string
		want          []string
	}{
		{"fix #b", "tui", []string{"fix #backend", "fix #bug"}},
		{"修正 >ops #b", "ops", []string{"修正 >ops #backend", "修正 >ops #bug"}},
		{"fix #bug #b", "tui", []string{"fix #bug #backend"}},
		{"#", "tui", []string{"#backend", "#bug"}},
		{"fix -- body #b", "", nil},
		{"fix #b ", "", nil},
		{"fix >missing #b", "", nil},
		{"fix !9 #b", "", nil},
	} {
		t.Run(tc.line, func(t *testing.T) {
			project, got := TagSuggestions(tc.line, utf8.RuneCountInString(tc.line), Prefixes{"tui", "ops"}, "tui", []string{"backend", "bug"})
			if project != tc.project || !slices.Equal(got, tc.want) {
				t.Fatalf("project=%q suggestions=%v; want %q %v", project, got, tc.project, tc.want)
			}
		})
	}
	if _, got := TagSuggestions("fix #b", 5, Prefixes{"tui"}, "tui", []string{"bug"}); len(got) != 0 {
		t.Fatalf("completion must not modify text after the cursor: %v", got)
	}
}

func TestTagSuggestionsExcludeNamesOutsideQuickAddGrammar(t *testing.T) {
	_, got := TagSuggestions("fix #", 5, Prefixes{"tui"}, "tui", []string{"two words", "bad\tname", "bad\nname", "!urgent", "", "bug:ui"})
	if !slices.Equal(got, []string{"fix #bug:ui"}) {
		t.Fatalf("completion must not change the title or introduce grammar tokens: %v", got)
	}
}
