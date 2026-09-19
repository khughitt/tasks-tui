package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

func TestAgo(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	for iso, want := range map[string]string{
		"2026-09-13T11:59:30Z": "now", "2026-09-13T11:15:00Z": "45m",
		"2026-09-13T09:00:00Z": "3h", "2026-09-10T12:00:00Z": "3d",
		"2026-07-01T12:00:00Z": "10w", "garbage": "?",
	} {
		if got := ago(iso, now); got != want {
			t.Errorf("ago(%q) = %q, want %q", iso, got, want)
		}
	}
}

func TestRowMarksSitInFixedCellsAndDetailsTrail(t *testing.T) {
	s := NewStyles(darkTone)
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	size := "m"
	full := fromRow(tasksctl.Row{ID: "tui-1", Title: "Do it", Status: "doing", Priority: 1, Size: &size, Updated: "2026-09-14T10:00:00Z", Tags: []string{"ui"}, OpenDescendantCount: 2,
		Claim: &tasksctl.ClaimInfo{Owner: "feat/x", Live: true}, Park: &tasksctl.ParkInfo{WaitingOn: "user"}, Periodic: &tasksctl.Periodic{Every: "30d"}}, 2)
	periodicOnly := fromRow(tasksctl.Row{ID: "tui-2", Title: "Sweep", Status: "todo", Priority: 3, Updated: "2026-09-10T10:00:00Z", Periodic: &tasksctl.Periodic{Every: "7d"}}, 2)
	rt := s.layoutRows([]rowView{full, periodicOnly}, 120, now)
	l0, l1 := ansi.Strip(rt.line(0, false)), ansi.Strip(rt.line(1, false))
	marks := taskTable.offset(rt.widths, "marks")
	if got := string([]rune(l0)[marks : marks+4]); got != "◆⏸⟳▸" {
		t.Fatalf("marks cell %q", got)
	}
	if got := string([]rune(l1)[marks : marks+4]); got != "  ⟳ " {
		t.Fatalf("periodic-only marks %q: ⟳ stays in the third cell", got)
	}
	for _, want := range []string{"tui-1", "P1", "m", "doing", "2h", "Do it", "[ui]", "feat/x", "user", "30d", "2 open"} {
		if !strings.Contains(l0, want) {
			t.Errorf("row lacks %q: %s", want, l0)
		}
	}
	if strings.Contains(l0, "◆ feat/x") || strings.Contains(l0, "⟳ 30d") {
		t.Fatalf("glyphs are no longer glued to their details: %s", l0)
	}
	if col(t, l0, "Do it") != col(t, l1, "Sweep") {
		t.Fatalf("titles must start at the same cell:\n%s\n%s", l0, l1)
	}
	if !strings.Contains(l1, "4d") {
		t.Fatalf("age: %s", l1)
	}
	styled := rt.line(0, false)
	doing := strings.TrimSuffix(s.Accent(2).Bold(true).Render("doing"), "\x1b[m")
	if !strings.Contains(styled, doing) {
		t.Errorf("doing lacks slot accent: %q", styled)
	}
	recent := strings.TrimSuffix(s.Recent.Render("2h"), "\x1b[m")
	old := strings.TrimSuffix(s.Muted.Render("4d"), "\x1b[m")
	if !strings.Contains(styled, recent) || !strings.Contains(rt.line(1, false), old) {
		t.Errorf("age fade: recent %q old %q", styled, rt.line(1, false))
	}
}

func TestSelectedRowCarriesSurfaceAndGutter(t *testing.T) {
	s := NewStyles(darkTone)
	rt := s.layoutRows([]rowView{fromRow(row("tui-1", "todo", 2), 2)}, 60, time.Now())
	sel, plain := rt.line(0, true), rt.line(0, false)
	if !strings.Contains(sel, "48;2;") || strings.Contains(plain, "48;2;") {
		t.Fatalf("surface only on the selected row:\n%q\n%q", sel, plain)
	}
	if !strings.HasPrefix(ansi.Strip(sel), "▌ ") || !strings.HasPrefix(ansi.Strip(plain), "  ") {
		t.Fatalf("gutter cursor: %q / %q", ansi.Strip(sel), ansi.Strip(plain))
	}
	if lipgloss.Width(sel) != 60 || lipgloss.Width(plain) != 60 {
		t.Fatalf("rows fill the width: %d %d", lipgloss.Width(sel), lipgloss.Width(plain))
	}
}

func TestStaleClaimAndUnresolvedRows(t *testing.T) {
	s := NewStyles(darkTone)
	r := fromRow(row("tui-1", "todo", 2), 2)
	r.Claim = &tasksctl.ClaimInfo{Owner: "feat/x", Live: false}
	p := tasksctl.ParkedRow{ID: "tui-2", Title: "Parked only", Park: &tasksctl.ParkInfo{WaitingOn: "agent", NextStep: "resume"}}
	parked := fromParked(p, 2)
	rt := s.layoutRows([]rowView{r, parked}, 120, time.Now())
	l0, l1 := ansi.Strip(rt.line(0, false)), ansi.Strip(rt.line(1, false))
	if !strings.Contains(l0, "◇") || strings.Contains(l0, "◆") || !strings.Contains(l0, "feat/x") {
		t.Errorf("stale claim: %s", l0)
	}
	if !strings.Contains(l1, "unresolved") || !strings.Contains(l1, "⏸") || !strings.Contains(l1, "agent") {
		t.Errorf("unresolved parked row: %s", l1)
	}
	if !parked.target().Unresolved {
		t.Fatal("unresolved rows produce unresolved targets")
	}
	if rt.header() == "" || !strings.Contains(ansi.Strip(rt.header()), "status") {
		t.Fatalf("header: %q", rt.header())
	}
}
