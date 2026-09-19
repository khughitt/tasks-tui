package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/khughitt/tasks-tui/internal/identity"
)

// Styles holds the semantic colours of spec §8.3 — the only hard-coded colours — and
// the per-slot accent styles derived from identity.
type Styles struct {
	Tone        identity.Tone
	Base        lipgloss.Style
	Muted       lipgloss.Style
	Bold        lipgloss.Style
	Recent      lipgloss.Style
	Frame       lipgloss.Style
	Header      lipgloss.Style
	Error       lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style
	ParkUser    lipgloss.Style
	ParkAgent   lipgloss.Style
	Claim       lipgloss.Style
	Stale       lipgloss.Style
	Periodic    lipgloss.Style
	Tag         lipgloss.Style
	cursorBlink bool
	priority    [5]lipgloss.Style
	status      map[string]lipgloss.Style
	accent      [identity.SlotCount]lipgloss.Style
	dim         [identity.SlotCount]lipgloss.Style
	surface     [identity.SlotCount]lipgloss.Style
	gutter      [identity.SlotCount]lipgloss.Style
	pill        [identity.SlotCount]lipgloss.Style
}

func NewStyles(t identity.Tone) *Styles {
	ld := lipgloss.LightDark(t.Dark)
	red := ld(lipgloss.Color("124"), lipgloss.Color("203"))
	orange := ld(lipgloss.Color("166"), lipgloss.Color("215"))
	yellow := ld(lipgloss.Color("136"), lipgloss.Color("221"))
	teal := ld(lipgloss.Color("30"), lipgloss.Color("80"))
	dim := ld(lipgloss.Color("245"), lipgloss.Color("243"))

	s := &Styles{Tone: t, cursorBlink: true}
	s.Base = lipgloss.NewStyle()
	s.Muted = lipgloss.NewStyle().Foreground(dim)
	s.Bold = lipgloss.NewStyle().Bold(true)
	s.Recent = lipgloss.NewStyle()
	s.Frame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(1, 2)
	pillText := ld(lipgloss.Color("#fafafa"), lipgloss.Color("#101010"))
	s.Header = lipgloss.NewStyle().Bold(true)
	s.Error = lipgloss.NewStyle().Foreground(red).Bold(true)
	s.Warning = lipgloss.NewStyle().Foreground(yellow)
	s.Info = s.Muted
	s.ParkUser = lipgloss.NewStyle().Foreground(yellow)
	s.ParkAgent = s.Muted
	s.Periodic = lipgloss.NewStyle().Foreground(teal)
	s.Stale = s.Muted
	s.priority = [5]lipgloss.Style{
		lipgloss.NewStyle().Foreground(red).Bold(true),
		lipgloss.NewStyle().Foreground(orange),
		lipgloss.NewStyle().Foreground(yellow),
		lipgloss.NewStyle(),
		s.Muted,
	}
	s.status = map[string]lipgloss.Style{
		"todo":    lipgloss.NewStyle(),
		"doing":   lipgloss.NewStyle().Bold(true),
		"blocked": lipgloss.NewStyle().Foreground(red),
		"idea":    lipgloss.NewStyle().Foreground(dim).Italic(true),
		"shelved": lipgloss.NewStyle().Foreground(dim).Italic(true),
		"done":    lipgloss.NewStyle().Foreground(dim).Strikethrough(true),
		"dropped": lipgloss.NewStyle().Foreground(dim).Strikethrough(true),
	}
	for slot := 0; slot < identity.SlotCount; slot++ {
		accent, surface := identity.Accent(slot, t), identity.Surface(slot, t)
		s.accent[slot] = lipgloss.NewStyle().Foreground(accent)
		s.dim[slot] = lipgloss.NewStyle().Foreground(identity.Dim(slot, t))
		s.surface[slot] = lipgloss.NewStyle().Background(surface)
		s.gutter[slot] = lipgloss.NewStyle().Foreground(accent).Background(surface)
		s.pill[slot] = lipgloss.NewStyle().Background(accent).Foreground(pillText).Bold(true)
	}
	s.Claim = s.Bold
	s.Tag = s.Muted
	return s
}

func (s *Styles) Accent(slot int) lipgloss.Style  { return s.accent[slot] }
func (s *Styles) Dim(slot int) lipgloss.Style     { return s.dim[slot] }
func (s *Styles) Surface(slot int) lipgloss.Style { return s.surface[slot] }
func (s *Styles) Gutter(slot int) lipgloss.Style  { return s.gutter[slot] }
func (s *Styles) Pill(slot int) lipgloss.Style    { return s.pill[slot] }
func (s *Styles) Rule(slot int) lipgloss.Style    { return s.accent[slot] }

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

// GlamourStyle is the standard style matching the tone (spec §8.2).
func (s *Styles) GlamourStyle() string {
	if s.Tone.Dark {
		return "dark"
	}
	return "light"
}
