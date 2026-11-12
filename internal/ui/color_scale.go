package ui

import (
	"image/color"
	"math"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func countPosition(n, maximum int) float64 {
	if n <= 0 {
		return 0
	}
	return min(1, math.Log1p(float64(n))/math.Log1p(float64(max(maximum, 4))))
}

func recencyPosition(iso string, now time.Time) (float64, bool) {
	stamp, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return 0, false
	}
	minutes := max(0, now.Sub(stamp).Minutes())
	return max(0, 1-math.Log1p(minutes)/math.Log1p(43200)), true
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	linear := func(v uint32) float64 {
		x := float64(v) / 65535
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
}

func contrastRatio(a, b color.Color) float64 {
	x, y := luminance(a), luminance(b)
	return (max(x, y) + 0.05) / (min(x, y) + 0.05)
}

// Adjust only generated data colors; authored identity and terminal endpoints stay exact.
func readableColor(c, bg color.Color, floor float64) color.Color {
	if contrastRatio(c, bg) >= floor {
		return c
	}
	var neutral color.Color = color.Black
	if contrastRatio(color.White, bg) > contrastRatio(color.Black, bg) {
		neutral = color.White
	}
	low, high := 0.0, 1.0
	for range 20 {
		midpoint := (low + high) / 2
		if contrastRatio(blendRGB(c, neutral, midpoint), bg) >= floor {
			high = midpoint
		} else {
			low = midpoint
		}
	}
	return blendRGB(c, neutral, high)
}

func (s *Styles) countStyle(n, maximum int, background color.Color) lipgloss.Style {
	if n == 0 {
		return s.Muted
	}
	if s.profile < colorprofile.ANSI256 || s.background == nil || background == nil {
		return s.Base
	}
	position := countPosition(n, maximum)
	blue, cyan, gold := lipgloss.Color("#4d9de0"), lipgloss.Color("#39b8c1"), lipgloss.Color("#e9b44c")
	var c color.Color
	if position < 0.5 {
		c = blendRGB(blue, cyan, 2*position)
	} else {
		c = blendRGB(cyan, gold, 2*position-1)
	}
	return s.Base.Foreground(readableColor(c, background, 4.5))
}

func (s *Styles) ageStyle(iso string, now time.Time, background color.Color) lipgloss.Style {
	position, ok := recencyPosition(iso, now)
	if !ok {
		return s.Muted
	}
	if s.profile < colorprofile.ANSI256 || s.foreground == nil || s.background == nil || background == nil {
		return s.Base
	}
	c := blendRGB(s.foreground, lipgloss.Color("#39b8c1"), position)
	floor := min(4.5, contrastRatio(s.foreground, background))
	return s.Base.Foreground(readableColor(c, background, floor))
}
