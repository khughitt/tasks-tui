package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"tasks-tui/internal/tasksctl"
)

type rowView struct {
	ID, Prefix, Title, Status, Size, Complexity, Process, Updated string
	Slot, Open                                                    int
	Priority                                                      *int
	Tags                                                          []string
	Claim                                                         *tasksctl.ClaimInfo
	Park                                                          *tasksctl.ParkInfo
	Periodic                                                      *tasksctl.Periodic
	Unresolved                                                    bool
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func fromRow(r tasksctl.Row, slot int) rowView {
	p := r.Priority
	return rowView{ID: r.ID, Prefix: r.Prefix(), Slot: slot, Title: r.Title, Status: r.Status, Priority: &p, Size: deref(r.Size), Complexity: deref(r.Complexity), Process: deref(r.Process), Updated: r.Updated, Tags: r.Tags, Open: r.OpenDescendantCount, Claim: r.Claim, Park: r.Park, Periodic: r.Periodic}
}
func fromParked(r tasksctl.ParkedRow, slot int) rowView {
	v := rowView{ID: r.ID, Prefix: r.Prefix(), Slot: slot, Title: r.Title, Status: deref(r.Status), Priority: r.Priority, Size: deref(r.Size), Complexity: deref(r.Complexity), Process: deref(r.Process), Updated: deref(r.Updated), Tags: r.Tags, Claim: r.Claim, Park: r.Park, Unresolved: r.Unresolved()}
	if r.OpenDescendantCount != nil {
		v.Open = *r.OpenDescendantCount
	}
	return v
}
func (r rowView) target() target {
	return target{ID: r.ID, Title: r.Title, Prefix: r.Prefix, Park: r.Park, Claim: r.Claim, Unresolved: r.Unresolved}
}
func ago(iso string, now time.Time) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return "?"
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + "m"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d.Hours())) + "h"
	case d < 14*24*time.Hour:
		return strconv.Itoa(int(d.Hours()/24)) + "d"
	default:
		return strconv.Itoa(int(d.Hours()/(24*7))) + "w"
	}
}
func pad(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

// taskTable is spec v1.1 §4: the columns of every task list.
var taskTable = table{gap: 2, flexMin: 24, cols: []column{
	{key: "gutter", min: 2},
	{key: "id", label: "id", min: 2},
	{key: "prio", label: "P", min: 2},
	{key: "size", label: "sz", min: 2, drop: 2},
	{key: "cx", label: "cx", min: 4, drop: 3},
	{key: "proc", label: "proc", min: 7, drop: 4},
	{key: "status", label: "status", min: 6},
	{key: "age", label: "age", align: alignRight, min: 3, drop: 1},
	{key: "marks", min: 4},
	{key: "title", label: "title", min: 5, flex: true},
}}

func (s *Styles) ageStyle(age string) lipgloss.Style {
	if strings.HasSuffix(age, "d") || strings.HasSuffix(age, "w") || age == "?" {
		return s.Muted
	}
	return s.Recent
}

func (s *Styles) taskCells(r rowView, now time.Time) []cell {
	accent := s.Accent(r.Slot)
	prio, prioStyle := "", s.Muted
	if r.Priority != nil {
		prio, prioStyle = "P"+strconv.Itoa(*r.Priority), s.PriorityStyle(*r.Priority)
	}
	status, statusStyle := r.Status, s.StatusStyle(r.Status)
	if r.Unresolved {
		status = "unresolved"
	}
	if r.Status == "doing" {
		statusStyle = statusStyle.Foreground(accent.GetForeground())
	}
	age := ago(r.Updated, now)
	marks := []span{{" ", s.Base}, {" ", s.Base}, {" ", s.Base}, {" ", s.Base}}
	var tail []span
	if len(r.Tags) > 0 {
		tail = append(tail, span{"[" + strings.Join(r.Tags, ", ") + "]", s.Tag})
	}
	if c := r.Claim; c != nil {
		style := s.Claim.Foreground(accent.GetForeground())
		glyph := "◆"
		if !c.Live {
			style, glyph = s.Stale, "◇"
		}
		marks[0] = span{glyph, style}
		tail = append(tail, span{c.Owner, style})
	}
	if p := r.Park; p != nil {
		style, who := s.ParkAgent, "agent"
		if p.WaitingOn == "user" {
			style, who = s.ParkUser, "user"
		}
		marks[1] = span{"⏸", style}
		tail = append(tail, span{who, style})
	}
	if r.Periodic != nil {
		marks[2] = span{"⟳", s.Periodic}
		tail = append(tail, span{r.Periodic.Every, s.Periodic})
	}
	if r.Open > 0 {
		marks[3] = span{"▸", s.Muted}
		tail = append(tail, span{fmt.Sprintf("%d open", r.Open), s.Muted})
	}
	return []cell{
		text(s.Base, "  "),
		text(accent, r.ID),
		text(prioStyle, prio),
		text(s.Muted, r.Size),
		text(s.Muted, r.Complexity),
		text(s.Muted, r.Process),
		text(statusStyle, status),
		text(s.ageStyle(age), age),
		{spans: marks},
		{spans: []span{{r.Title, statusStyle}}, tail: tail},
	}
}

type rowTable struct {
	s         *Styles
	rows      []rowView
	cells     [][]cell
	widths    []int
	width     int
	labels    map[string]string
	candidate string
	slot      int
}

func (s *Styles) layoutRows(rows []rowView, width int, now time.Time) rowTable {
	return s.layoutRowsWithLabels(rows, width, now, nil)
}

func (s *Styles) layoutRowsWithLabels(rows []rowView, width int, now time.Time, labels map[string]string) rowTable {
	rt := rowTable{s: s, rows: rows, width: width, cells: make([][]cell, len(rows))}
	for i, r := range rows {
		rt.cells[i] = s.taskCells(r, now)
	}
	rt.widths = taskTable.widthsWithLabels(rt.cells, width, labels)
	rt.labels = labels
	return rt
}

func (rt rowTable) header() string {
	return fit(taskTable.headerWithCandidate(rt.widths, rt.s, rt.labels, rt.candidate, rt.slot), rt.width, lipgloss.NewStyle())
}

func (rt rowTable) line(i int, selected bool) string {
	cells, base := rt.cells[i], lipgloss.NewStyle()
	if selected {
		slot := rt.rows[i].Slot
		cells = append([]cell{text(rt.s.Gutter(slot), "▌ ")}, cells[1:]...)
		base = rt.s.Surface(slot)
	}
	return fit(taskTable.row(rt.widths, cells, base), rt.width, base)
}

func (rt rowTable) titleOffset() int { return max(0, taskTable.offset(rt.widths, "title")) }
