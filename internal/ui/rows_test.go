package ui

import (
	"strings"
	"testing"
	"time"

	"tasks-tui/internal/identity"
	"tasks-tui/internal/tasksctl"
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

func TestRenderRowMarkersAndSlotAccents(t *testing.T) {
	s := NewStyles(identity.Tone{Dark: true, SatScale: 1})
	size := "m"
	r := tasksctl.Row{ID: "tui-1", Title: "Do it", Status: "doing", Priority: 1, Size: &size, Updated: "2026-09-13T10:00:00Z", Tags: []string{"ui"}, OpenDescendantCount: 2,
		Claim: &tasksctl.ClaimInfo{Owner: "feat/x", Live: true}, Park: &tasksctl.ParkInfo{WaitingOn: "user"}, Periodic: &tasksctl.Periodic{Every: "30d"}}
	out := s.renderRow(fromRow(r, 2), 120, false)
	for _, want := range []string{"tui-1", "P1", "m", "doing", "Do it", "[ui]", "◆ feat/x", "⏸ user", "⟳ 30d", "▸ 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("row lacks %q: %s", want, out)
		}
	}
	doing := strings.TrimSuffix(s.Accent(2).Bold(true).Render("doing"), "\x1b[m")
	claim := strings.TrimSuffix(s.Accent(2).Bold(true).Render("◆ feat/x"), "\x1b[m")
	if !strings.Contains(out, doing) || !strings.Contains(out, claim) {
		t.Errorf("doing and live claim lack slot accent: %q", out)
	}
	r.Claim.Live = false
	out = s.renderRow(fromRow(r, 2), 120, false)
	if !strings.Contains(out, "◇ feat/x") || strings.Contains(out, "◆") {
		t.Errorf("stale claim marker wrong: %s", out)
	}
	p := tasksctl.ParkedRow{ID: "tui-2", Title: "Parked only", Park: &tasksctl.ParkInfo{WaitingOn: "agent", NextStep: "resume"}}
	out = s.renderRow(fromParked(p, 2), 120, false)
	if !strings.Contains(out, "unresolved") || !strings.Contains(out, "⏸ agent") {
		t.Errorf("unresolved row: %s", out)
	}
	if !fromParked(p, 2).target().Unresolved {
		t.Fatal("unresolved rows produce unresolved targets")
	}
}
