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
func (s *Styles) renderRow(r rowView, width int, selected bool) string {
	accent := s.Accent(r.Slot)
	prio, prioStyle := "  ", s.Muted
	if r.Priority != nil {
		prio = "P" + strconv.Itoa(*r.Priority)
		prioStyle = s.PriorityStyle(*r.Priority)
	}
	status := r.Status
	if r.Unresolved {
		status = "unresolved"
	}
	statusStyle := s.StatusStyle(r.Status)
	if r.Status == "doing" {
		statusStyle = statusStyle.Foreground(accent.GetForeground())
	}
	cols := []string{accent.Render(pad(r.ID, 12)), prioStyle.Render(prio), s.Muted.Render(pad(r.Size, 2)), s.Muted.Render(pad(r.Complexity, 4)), s.Muted.Render(pad(r.Process, 7)), statusStyle.Render(pad(status, 10)), s.Muted.Render(pad(ago(r.Updated, time.Now()), 4))}
	line := strings.Join(cols, " ") + " " + statusStyle.Render(r.Title)
	var marks []string
	if len(r.Tags) > 0 {
		marks = append(marks, s.Tag.Render("["+strings.Join(r.Tags, ", ")+"]"))
	}
	if c := r.Claim; c != nil {
		if c.Live {
			marks = append(marks, s.Claim.Foreground(accent.GetForeground()).Render("◆ "+c.Owner))
		} else {
			marks = append(marks, s.Stale.Render("◇ "+c.Owner))
		}
	}
	if p := r.Park; p != nil {
		if p.WaitingOn == "user" {
			marks = append(marks, s.ParkUser.Render("⏸ user"))
		} else {
			marks = append(marks, s.ParkAgent.Render("⏸ agent"))
		}
	}
	if r.Periodic != nil {
		marks = append(marks, s.Periodic.Render("⟳ "+r.Periodic.Every))
	}
	if r.Open > 0 {
		marks = append(marks, s.Muted.Render(fmt.Sprintf("▸ %d", r.Open)))
	}
	if len(marks) > 0 {
		line += "  " + strings.Join(marks, " ")
	}
	st := lipgloss.NewStyle().MaxWidth(width)
	if selected {
		st = st.Reverse(true)
	}
	return st.Render(pad(line, width))
}
