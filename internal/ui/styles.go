package ui

import (
	"image/color"
	"math"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

// Styles owns terminal roles and authored project colors; unknown colors stay unset.
type Styles struct {
	Base, Muted, Bold, Recent, Frame, Header  lipgloss.Style
	Error, Warning, Info, ParkUser, ParkAgent lipgloss.Style
	Claim, Stale, Periodic, Tag               lipgloss.Style
	cursorBlink                               bool
	priority                                  [5]lipgloss.Style
	status                                    map[string]lipgloss.Style
	profile                                   colorprofile.Profile
	background, foreground                    color.Color
	projects                                  map[string]tasksctl.Project
}

func NewStyles() *Styles {
	s := &Styles{projects: map[string]tasksctl.Project{}, cursorBlink: true}
	s.setProfile(colorprofile.Unknown)
	return s
}

func (s *Styles) setProfile(p colorprofile.Profile) {
	s.profile = p
	s.Base = lipgloss.NewStyle()
	s.Bold = s.Base.Bold(true)
	s.Muted = s.Base.Faint(true).Bold(false)
	s.Recent, s.Header = s.Base, s.Bold
	s.Frame = s.Base.Border(lipgloss.RoundedBorder()).Padding(1, 2)
	s.Error, s.Warning, s.Periodic = s.Bold, s.Base, s.Base
	if p >= colorprofile.ANSI {
		s.Error = s.Error.Foreground(lipgloss.Red)
		s.Warning = s.Warning.Foreground(lipgloss.Yellow)
		s.Periodic = s.Periodic.Foreground(lipgloss.Cyan)
	}
	s.Info, s.ParkUser, s.ParkAgent = s.Base, s.Warning, s.Base
	s.Claim, s.Stale, s.Tag = s.Bold, s.Muted, s.Base
	s.priority = [5]lipgloss.Style{s.Error, s.Warning.Bold(true), s.Warning, s.Base, s.Muted}
	s.status = map[string]lipgloss.Style{
		"todo": s.Base, "doing": s.Bold, "blocked": s.Error.UnsetBold(),
		"idea": s.Base.Italic(true), "shelved": s.Muted.Italic(true),
		"done": s.Muted.Strikethrough(true), "dropped": s.Muted.Strikethrough(true),
	}
}

func (s *Styles) projectColor(prefix string) color.Color {
	p, ok := s.projects[prefix]
	if !ok || p.Color == nil {
		return nil
	}
	switch s.profile {
	case colorprofile.TrueColor:
		return lipgloss.Color(*p.Color)
	case colorprofile.ANSI256:
		if p.Color256 != nil {
			return lipgloss.ANSIColor(*p.Color256)
		}
	}
	return nil
}

func (s *Styles) Accent(prefix string) lipgloss.Style {
	if c := s.projectColor(prefix); c != nil {
		return s.Base.Foreground(c)
	}
	return s.Base
}

func blendRGB(a, b color.Color, fraction float64) color.Color {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	mix := func(x, y uint32) uint8 { return uint8(math.Round((1-fraction)*float64(x>>8) + fraction*float64(y>>8))) }
	return color.RGBA{R: mix(ar, br), G: mix(ag, bg), B: mix(ab, bb), A: 255}
}

func (s *Styles) rowBackground(prefix string, selected bool) color.Color {
	if selected && s.profile == colorprofile.TrueColor && s.background != nil {
		if c := s.projectColor(prefix); c != nil {
			return blendRGB(s.background, c, 0.12)
		}
	}
	return s.background
}

func (s *Styles) Surface(prefix string) lipgloss.Style {
	if s.profile == colorprofile.TrueColor && s.background != nil && s.projectColor(prefix) != nil {
		return s.Bold.Background(s.rowBackground(prefix, true))
	}
	return s.Bold
}

func (s *Styles) Gutter(prefix string) lipgloss.Style {
	return s.Accent(prefix).Inherit(s.Surface(prefix))
}
func (s *Styles) Rule(prefix string) lipgloss.Style { return s.Accent(prefix) }
func (s *Styles) ActiveLabel(prefix, label string) string {
	if s.projectColor(prefix) == nil {
		label = "[" + label + "]"
	}
	return s.Accent(prefix).Bold(true).Underline(true).Render(label)
}

func (s *Styles) PriorityStyle(p int) lipgloss.Style {
	if p < 0 || p > 4 {
		return s.Base
	}
	return s.priority[p]
}
func (s *Styles) StatusStyle(status string) lipgloss.Style {
	if st, ok := s.status[status]; ok {
		return st
	}
	return s.Base
}

func (s *Styles) inputStyles() textinput.Styles {
	state := textinput.StyleState{Text: s.Base, Prompt: s.Base, Suggestion: s.Base, Placeholder: s.Muted}
	return textinput.Styles{Focused: state, Blurred: state}
}
func (s *Styles) textareaStyles() textarea.Styles {
	state := textarea.StyleState{Base: s.Base, Text: s.Base, Prompt: s.Base, CursorLine: s.Base,
		LineNumber: s.Base, CursorLineNumber: s.Bold, Placeholder: s.Muted, EndOfBuffer: s.Muted,
		Selection: s.Bold.Underline(true)}
	return textarea.Styles{Focused: state, Blurred: state}
}
