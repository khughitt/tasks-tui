package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type legendGroup struct {
	title string
	keys  []key.Binding
}

func renderGroup(s *Styles, g legendGroup) string {
	kw := 0
	for _, b := range g.keys {
		kw = max(kw, lipgloss.Width(b.Help().Key))
	}
	lines := []string{s.Muted.Render(g.title)}
	for _, b := range g.keys {
		h := b.Help()
		lines = append(lines, s.Bold.Render(strings.Repeat(" ", kw-lipgloss.Width(h.Key))+h.Key)+"  "+h.Desc)
	}
	return strings.Join(lines, "\n")
}

// legendContent lays the groups out four across, else two by two, else stacked.
func legendContent(s *Styles, width int) string {
	groups := make([]string, len(legendGroups))
	for i, g := range legendGroups {
		groups[i] = renderGroup(s, g)
	}
	across := func(gs ...string) string {
		parts := make([]string, 0, 2*len(gs))
		for i, g := range gs {
			if i > 0 {
				parts = append(parts, "    ")
			}
			parts = append(parts, g)
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	}
	if four := across(groups...); lipgloss.Width(four) <= width {
		return four
	}
	if two := lipgloss.JoinVertical(lipgloss.Left, across(groups[0], groups[1]), "", across(groups[2], groups[3])); lipgloss.Width(two) <= width {
		return two
	}
	return lipgloss.JoinVertical(lipgloss.Left, groups[0], "", groups[1], "", groups[2], "", groups[3])
}
