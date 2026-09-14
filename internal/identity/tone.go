package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"

	"charm.land/lipgloss/v2"
)

// Tone is familiar's scheme.json: the scheme owns lightness, the identity owns hue.
type Tone struct {
	Dark     bool
	SatScale float64
}

// LoadTone reads ~/.config/familiar/scheme.json; missing is dark with scale 1.
func LoadTone(path string) (Tone, error) {
	text, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Tone{Dark: true, SatScale: 1}, nil
	}
	if err != nil {
		return Tone{}, err
	}
	var raw struct {
		Mode     string   `json:"mode"`
		SatScale *float64 `json:"satScale"`
	}
	if err := json.Unmarshal(text, &raw); err != nil {
		return Tone{}, fmt.Errorf("scheme.json: %w", err)
	}
	t := Tone{SatScale: 1}
	switch raw.Mode {
	case "dark":
		t.Dark = true
	case "light":
		t.Dark = false
	default:
		return Tone{}, fmt.Errorf("scheme.json: mode must be dark or light, got %q", raw.Mode)
	}
	if raw.SatScale != nil {
		if *raw.SatScale < 0 || *raw.SatScale > 2 {
			return Tone{}, fmt.Errorf("scheme.json: satScale must be in 0..2, got %v", *raw.SatScale)
		}
		t.SatScale = *raw.SatScale
	}
	return t, nil
}

// familiar's ramp anchors (src/theme/ramp.js): base is the accent, shadow the dim variant.
type anchors struct{ shadow, base float64 }

func anchorsFor(t Tone) anchors {
	if t.Dark {
		return anchors{shadow: 36, base: 58}
	}
	return anchors{shadow: 30, base: 46}
}

func Accent(slot int, t Tone) color.Color {
	hs := Slots[slot]
	return lipgloss.Color(HSLToHex(float64(hs.Hue), math.Max(0, math.Min(float64(hs.Sat)*t.SatScale, 100)), anchorsFor(t).base))
}

func Dim(slot int, t Tone) color.Color {
	hs := Slots[slot]
	return lipgloss.Color(HSLToHex(float64(hs.Hue), math.Max(0, math.Min(float64(hs.Sat)*t.SatScale, 100)), anchorsFor(t).shadow))
}

// HSLToHex is familiar's hslToHex: h in degrees, s and l in percent.
func HSLToHex(h, s, l float64) string {
	S, L := s/100, l/100
	k := func(n float64) float64 { return math.Mod(n+h/30, 12) }
	a := S * math.Min(L, 1-L)
	f := func(n float64) float64 {
		return L - a*math.Max(-1, math.Min(k(n)-3, math.Min(9-k(n), 1)))
	}
	channel := func(v float64) int { return int(math.Round(255 * v)) }
	return fmt.Sprintf("#%02x%02x%02x", channel(f(0)), channel(f(8)), channel(f(4)))
}
