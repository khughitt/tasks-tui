package ui

import (
	"strings"
	"testing"

	"github.com/khughitt/tasks-tui/internal/identity"
)

func TestSlotStylesCarrySurfaceAndPill(t *testing.T) {
	s := NewStyles(darkTone)
	if out := s.Surface(3).Render("x"); !strings.Contains(out, "\x1b[48;2;") {
		t.Fatalf("surface is a background: %q", out)
	}
	if out := s.Gutter(3).Render("▌"); !strings.Contains(out, "\x1b[38;2;") || !strings.Contains(out, "48;2;") {
		t.Fatalf("gutter is accent over surface: %q", out)
	}
	pill := s.Pill(3).Render("1 Ready")
	if !strings.Contains(pill, "48;2;") || !strings.Contains(pill, "38;2;16;16;16") {
		t.Fatalf("pill is accent background with #101010 text on a dark tone: %q", pill)
	}
	lightPill := NewStyles(identity.Tone{Dark: false, SatScale: 1}).Pill(3).Render("x")
	if !strings.Contains(lightPill, "38;2;250;250;250") {
		t.Fatalf("pill text is #fafafa on a light tone: %q", lightPill)
	}
}
