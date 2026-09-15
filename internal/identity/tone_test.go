package identity

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestHSLToHex(t *testing.T) {
	if got := HSLToHex(0, 0, 100); got != "#ffffff" {
		t.Fatalf("white %s", got)
	}
	if got := HSLToHex(0, 100, 50); got != "#ff0000" {
		t.Fatalf("red %s", got)
	}
	if got := HSLToHex(120, 100, 25); got != "#008000" {
		t.Fatalf("green %s", got)
	}
}

func TestLoadTone(t *testing.T) {
	dir := t.TempDir()
	tone, err := LoadTone(filepath.Join(dir, "missing.json"))
	if err != nil || !tone.Dark || tone.SatScale != 1 {
		t.Fatalf("default tone %+v err=%v", tone, err)
	}
	p := filepath.Join(dir, "scheme.json")
	if err := os.WriteFile(p, []byte(`{"mode":"light","satScale":0.5}`), 0o644); err != nil {
		t.Fatal(err)
	}
	tone, err = LoadTone(p)
	if err != nil || tone.Dark || tone.SatScale != 0.5 {
		t.Fatalf("tone %+v err=%v", tone, err)
	}
	if err := os.WriteFile(p, []byte(`{"mode":"sepia"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadTone(p); err == nil {
		t.Fatal("unknown mode must error")
	}
}

func TestAccentUsesRampAnchors(t *testing.T) {
	dark := Accent(0, Tone{Dark: true, SatScale: 1})
	light := Accent(0, Tone{Dark: false, SatScale: 1})
	if dark != lipgloss.Color(HSLToHex(22, 62, 58)) || light != lipgloss.Color(HSLToHex(22, 62, 46)) {
		t.Fatalf("dark=%v light=%v", dark, light)
	}
	if Dim(0, Tone{Dark: true, SatScale: 1}) != lipgloss.Color(HSLToHex(22, 62, 36)) {
		t.Fatal("dim must use the shadow anchor")
	}
	if Accent(0, Tone{Dark: true, SatScale: 0.5}) != lipgloss.Color(HSLToHex(22, 31, 58)) {
		t.Fatal("satScale must scale saturation")
	}
	if Accent(0, Tone{Dark: true, SatScale: 2}) != lipgloss.Color(HSLToHex(22, 100, 58)) {
		t.Fatal("satScale must clamp saturation to 100")
	}
}

func TestSurfaceIsLowLightnessHalfSaturation(t *testing.T) {
	dark := Surface(0, Tone{Dark: true, SatScale: 1})
	light := Surface(0, Tone{Dark: false, SatScale: 1})
	if fmt.Sprint(dark) != fmt.Sprint(lipgloss.Color(HSLToHex(22, 31, 20))) {
		t.Fatalf("dark surface %v", dark)
	}
	if fmt.Sprint(light) != fmt.Sprint(lipgloss.Color(HSLToHex(22, 31, 92))) {
		t.Fatalf("light surface %v", light)
	}
}
