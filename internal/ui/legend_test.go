package ui

import (
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestEveryGroupedBindingHasHelp(t *testing.T) {
	for _, g := range legendGroups {
		for _, b := range g.keys {
			if h := b.Help(); h.Key == "" || h.Desc == "" {
				t.Errorf("%s: binding %v lacks help", g.title, b.Keys())
			}
		}
	}
	if len(legendGroups) != 4 {
		t.Fatalf("%d groups", len(legendGroups))
	}
}

func TestLegendGroupsReflowFourAcrossTwoByTwoThenStacked(t *testing.T) {
	s := coloredStyles()
	wide := ansi.Strip(legendGroupsContent(s, 200))
	if lipgloss.Height(wide) > 9 {
		t.Fatalf("four across is one band:\n%s", wide)
	}
	mid := ansi.Strip(legendGroupsContent(s, 60))
	if h := lipgloss.Height(mid); h <= lipgloss.Height(wide) || h > 20 {
		t.Fatalf("two by two height %d:\n%s", h, mid)
	}
	narrow := ansi.Strip(legendGroupsContent(s, 24))
	if lipgloss.Height(narrow) <= lipgloss.Height(mid) {
		t.Fatalf("stacked is tallest:\n%s", narrow)
	}
	for _, want := range []string{"j/k", "move", "h/l", "enter/i", "open", "c …", "launch agent", "space", "F5", "g g/G", "quit / close", "1–5"} {
		if !strings.Contains(wide, want) {
			t.Errorf("legend lacks %q", want)
		}
	}
}

func TestLegendSortBoxListsEachViewsColumns(t *testing.T) {
	box := ansi.Strip(renderSortBox(coloredStyles()))
	want := []string{
		"sort  projects  tasks",
		"s p   prefix    priority",
		"s a   activity  age",
		"s t             title",
		"S …   descending",
	}
	if got := strings.Split(box, "\n"); !slices.Equal(got, want) {
		t.Fatalf("sort box:\n%s\nwant:\n%s", box, strings.Join(want, "\n"))
	}
}

func TestLegendGroupTitlesAlignWithTheirKeys(t *testing.T) {
	lines := strings.Split(ansi.Strip(legendGroupsContent(coloredStyles(), 200)), "\n")
	for title, first := range map[string]string{"move ": "j/k ", "open ": "enter/i ", "task ": "a ", "app ": "?  keys"} {
		if a, b := strings.Index(lines[0], title), strings.Index(lines[1], first); a < 0 || a != b {
			t.Errorf("%q at %d, its first key %q at %d:\n%s", title, a, first, b, strings.Join(lines[:2], "\n"))
		}
	}
}
