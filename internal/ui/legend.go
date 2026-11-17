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

// renderGroup left-aligns the group's keys under its title, descriptions in a column after them.
func renderGroup(s *Styles, g legendGroup) string {
	kw := 0
	for _, b := range g.keys {
		kw = max(kw, lipgloss.Width(b.Help().Key))
	}
	lines := []string{s.Base.Render(g.title)}
	for _, b := range g.keys {
		h := b.Help()
		lines = append(lines, s.Bold.Render(h.Key)+strings.Repeat(" ", kw-lipgloss.Width(h.Key))+"  "+h.Desc)
	}
	return strings.Join(lines, "\n")
}

// renderSortBox lays the sort chords out like a group: the title over the key column, one
// row per ascending chord with a column per view (blank where a view has no column under
// that key), then the descending prefix.
func renderSortBox(s *Styles) string {
	var seqs []string
	cols := make([]map[string]string, len(legendSortTables))
	widths := make([]int, len(legendSortTables))
	for i, t := range legendSortTables {
		cols[i] = map[string]string{}
		widths[i] = lipgloss.Width(t.view)
		for _, c := range sortChoices(t.table) {
			seq := "s " + c.key
			if !slices.Contains(seqs, seq) {
				seqs = append(seqs, seq)
			}
			cols[i][seq] = c.column
			widths[i] = max(widths[i], lipgloss.Width(c.column))
		}
	}
	const title, desc = "sort", "S …"
	kw := lipgloss.Width(title)
	for _, seq := range append([]string{desc}, seqs...) {
		kw = max(kw, lipgloss.Width(seq))
	}
	row := func(lead string, cells []string) string {
		parts := []string{lead + strings.Repeat(" ", kw-lipgloss.Width(lead))}
		for i, cell := range cells {
			parts = append(parts, cell+strings.Repeat(" ", max(0, widths[i]-lipgloss.Width(cell))))
		}
		return strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	views := make([]string, len(legendSortTables))
	for i, t := range legendSortTables {
		views[i] = s.Base.Render(t.view)
	}
	lines := []string{row(s.Base.Render(title), views)}
	for _, seq := range seqs {
		cells := make([]string, len(legendSortTables))
		for i := range legendSortTables {
			cells[i] = cols[i][seq]
		}
		lines = append(lines, row(s.Bold.Render(seq), cells))
	}
	lines = append(lines, row(s.Bold.Render(desc), []string{"descending"}))
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
