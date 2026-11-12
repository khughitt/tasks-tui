package ui

import (
	"github.com/charmbracelet/colorprofile"
	"strings"
	"testing"
)

func TestSurfaceAndGutterRequireTrueColorBackground(t *testing.T) {
	s := coloredStyles()
	if out := s.Surface("tui").Render("x"); !strings.Contains(out, "48;2;") {
		t.Fatalf("missing known tint: %q", out)
	}
	if out := s.Gutter("tui").Render("▌"); !strings.Contains(out, "38;2;18;58;188") || !strings.Contains(out, "48;2;") {
		t.Fatalf("gutter: %q", out)
	}
	s.setProfile(colorprofile.ANSI256)
	if out := s.Gutter("tui").Render("▌"); strings.Contains(out, "48;") || !strings.Contains(out, "38;5;42") {
		t.Fatalf("indexed gutter must use supplied index without tint: %q", out)
	}
	s.setProfile(colorprofile.TrueColor)
	s.background = nil
	if out := s.Surface("tui").Render("x"); strings.Contains(out, "48;") {
		t.Fatalf("guessed background: %q", out)
	}
}
