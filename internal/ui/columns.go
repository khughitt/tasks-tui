package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type align int

const (
	alignLeft align = iota
	alignRight
)

type column struct {
	key, label string
	align      align
	min        int
	flex       bool
	drop       int
}

type span struct {
	text  string
	style lipgloss.Style
}

type cell struct {
	spans []span
	tail  []span
}

func text(style lipgloss.Style, s string) cell { return cell{spans: []span{{s, style}}} }

func (c cell) width() int {
	w := 0
	for _, sp := range c.spans {
		w += lipgloss.Width(sp.text)
	}
	return w
}

type table struct {
	cols    []column
	gap     int
	flexMin int
}

func (t table) measure(rows [][]cell) []int {
	widths := make([]int, len(t.cols))
	for i, c := range t.cols {
		if c.flex {
			continue
		}
		widths[i] = c.min
		for _, r := range rows {
			if w := r[i].width(); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func (t table) sum(widths []int) int {
	total, n := 0, 0
	for i, c := range t.cols {
		switch {
		case c.flex:
			total += t.flexMin
		case widths[i] == 0:
			continue
		default:
			total += widths[i]
		}
		n++
	}
	return total + t.gap*max(0, n-1)
}

func (t table) minWidth(rows [][]cell) int {
	widths := t.measure(rows)
	for i, c := range t.cols {
		if c.drop != 0 {
			widths[i] = 0
		}
	}
	return t.sum(widths)
}

func (t table) widths(rows [][]cell, width int) []int {
	return t.widthsWithLabels(rows, width, nil)
}

func (t table) widthsWithLabels(rows [][]cell, width int, labels map[string]string) []int {
	widths := t.measure(rows)
	for i, c := range t.cols {
		if label := labels[c.key]; !c.flex && lipgloss.Width(label) > widths[i] {
			widths[i] = lipgloss.Width(label)
		}
	}
	for t.sum(widths) > width {
		best := -1
		for i, c := range t.cols {
			if c.flex || c.drop == 0 || widths[i] == 0 {
				continue
			}
			if best < 0 || c.drop > t.cols[best].drop {
				best = i
			}
		}
		if best < 0 {
			break
		}
		widths[best] = 0
	}
	for i, c := range t.cols {
		if c.flex {
			widths[i] = max(0, width-(t.sum(widths)-t.flexMin))
		}
	}
	return widths
}

//lint:ignore U1000 used by subsequent table renderers
func (t table) total(widths []int) int {
	total, n := 0, 0
	for _, w := range widths {
		if w > 0 {
			total += w
			n++
		}
	}
	return total + t.gap*max(0, n-1)
}

func (t table) offset(widths []int, key string) int {
	off := 0
	for i, c := range t.cols {
		if widths[i] == 0 {
			continue
		}
		if c.key == key {
			return off
		}
		off += widths[i] + t.gap
	}
	return -1
}

func (t table) header(widths []int, s *Styles) string {
	return t.headerWithLabels(widths, s, nil)
}

func (t table) headerWithLabels(widths []int, s *Styles, labels map[string]string) string {
	cells := make([]cell, len(t.cols))
	for i, c := range t.cols {
		label := c.label
		if labels[c.key] != "" {
			label = labels[c.key]
		}
		cells[i] = text(s.Muted, label)
	}
	return t.row(widths, cells, lipgloss.NewStyle())
}

func (t table) row(widths []int, cells []cell, base lipgloss.Style) string {
	gap := base.Render(strings.Repeat(" ", t.gap))
	var parts []string
	for i, c := range t.cols {
		w := widths[i]
		if w == 0 {
			continue
		}
		var s string
		if c.flex {
			s = t.flexText(cells[i], w, base)
		} else {
			s = spansText(cells[i].spans, base)
		}
		fill := base.Render(strings.Repeat(" ", max(0, w-lipgloss.Width(s))))
		if c.align == alignRight {
			s = fill + s
		} else {
			s += fill
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, gap)
}

func spansText(spans []span, base lipgloss.Style) string {
	var b strings.Builder
	for _, sp := range spans {
		b.WriteString(sp.style.Inherit(base).Render(sp.text))
	}
	return b.String()
}

func clip(s string, w int, base lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w-1, "") + base.Render("…")
}

func (t table) flexText(c cell, w int, base lipgloss.Style) string {
	gap := base.Render(strings.Repeat(" ", t.gap))
	head, headW := spansText(c.spans, base), c.width()
	pieces := make([]string, len(c.tail))
	for i, sp := range c.tail {
		pieces[i] = sp.style.Inherit(base).Render(sp.text)
	}
	rest := strings.Join(pieces, gap)
	restW := 0
	if rest != "" {
		restW = lipgloss.Width(rest) + t.gap
	}
	if headW+restW > w {
		keep := min(headW, max(w-restW, min(t.flexMin, w)))
		if keep < headW {
			head, headW = clip(head, keep, base), keep
		}
		if room := w - headW - t.gap; rest != "" && lipgloss.Width(rest) > room {
			rest = clip(rest, room, base)
		}
	}
	if rest == "" {
		return head
	}
	return head + gap + rest
}

func fit(line string, width int, base lipgloss.Style) string {
	if w := lipgloss.Width(line); w < width {
		return line + base.Render(strings.Repeat(" ", width-w))
	}
	return ansi.Truncate(line, width, "")
}
