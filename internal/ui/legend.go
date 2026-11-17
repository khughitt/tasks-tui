package ui

import (
	"slices"
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
	lines := []string{s.Base.Render(g.title)}
	for _, b := range g.keys {
		h := b.Help()
		lines = append(lines, s.Bold.Render(strings.Repeat(" ", kw-lipgloss.Width(h.Key))+h.Key)+"  "+h.Desc)
	}
	return strings.Join(lines, "\n")
}

// renderSortBox lists the sort chords: one row per second key, one column per view, blank
// where a view has no column under that key.
func renderSortBox(s *Styles) string {
	var letters []string
	cols := make([]map[string]string, len(legendSortTables))
	widths := make([]int, len(legendSortTables))
	for i, t := range legendSortTables {
		cols[i] = map[string]string{}
		widths[i] = lipgloss.Width(t.view)
		for _, c := range sortChoices(t.table) {
			if !slices.Contains(letters, c.key) {
				letters = append(letters, c.key)
			}
			cols[i][c.key] = c.column
			widths[i] = max(widths[i], lipgloss.Width(c.column))
		}
	}
	row := func(lead string, cells func(i int) string) string {
		parts := []string{lead}
		for i := range legendSortTables {
			cell := cells(i)
			parts = append(parts, cell+strings.Repeat(" ", widths[i]-lipgloss.Width(cell)))
		}
		return strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	lines := []string{
		s.Base.Render("sort") + "  " + s.Bold.Render("s") + " ascending · " + s.Bold.Render("S") + " descending, then",
		row(" ", func(i int) string { return s.Base.Render(legendSortTables[i].view) }),
	}
	for _, l := range letters {
		lines = append(lines, row(s.Bold.Render(l), func(i int) string { return cols[i][l] }))
	}
	return strings.Join(lines, "\n")
}

// legendContent lays the groups out four across, else two by two, else stacked, with
// the sort box below.
func legendContent(s *Styles, width int) string {
	return lipgloss.JoinVertical(lipgloss.Left, legendGroupsContent(s, width), "", renderSortBox(s))
}

func legendGroupsContent(s *Styles, width int) string {
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
