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

func TestLegendReflowsFourAcrossTwoByTwoThenStacked(t *testing.T) {
	s := NewStyles(darkTone)
	wide := ansi.Strip(legendContent(s, 200))
	if lipgloss.Height(wide) > 9 {
		t.Fatalf("four across is one band:\n%s", wide)
	}
	mid := ansi.Strip(legendContent(s, 60))
	if h := lipgloss.Height(mid); h <= lipgloss.Height(wide) || h > 20 {
		t.Fatalf("two by two height %d:\n%s", h, mid)
	}
	narrow := ansi.Strip(legendContent(s, 24))
	if lipgloss.Height(narrow) <= lipgloss.Height(mid) {
		t.Fatalf("stacked is tallest:\n%s", narrow)
	}
	for _, want := range []string{"j/k", "move", "enter/i", "open", "launch agent", "quit / close", "1–5"} {
		if !strings.Contains(wide, want) {
			t.Errorf("legend lacks %q", want)
		}
	}
}
