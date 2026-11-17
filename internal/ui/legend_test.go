package ui

import (
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
	box := ansi.Strip(legendContent(coloredStyles(), 200))
	for _, want := range []string{"s ascending", "S descending", "projects", "tasks", "prefix", "activity", "priority", "age", "title"} {
		if !strings.Contains(box, want) {
			t.Errorf("sort box lacks %q:\n%s", want, box)
		}
	}
	rows := map[string]string{}
	for _, line := range strings.Split(box, "\n") {
		if f := strings.Fields(line); len(f) >= 2 && len(f[0]) == 1 {
			rows[f[0]] = strings.Join(f[1:], " ")
		}
	}
	for letter, want := range map[string]string{"p": "prefix priority", "a": "activity age", "t": "title"} {
		if rows[letter] != want {
			t.Errorf("sort row %s = %q, want %q:\n%s", letter, rows[letter], want, box)
		}
	}
}
