package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

type stylesChangedMsg struct{}
type colorQueryTimeoutMsg struct{ generation uint64 }

func (a *App) refreshStyles() tea.Cmd {
	var cmds []tea.Cmd
	for i, v := range a.stack {
		var cmd tea.Cmd
		a.stack[i], cmd = v.update(stylesChangedMsg{})
		cmds = append(cmds, cmd)
	}
	if a.overlay != nil {
		var cmd tea.Cmd
		a.overlay, cmd = a.overlay.update(stylesChangedMsg{})
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (a *App) updateTheme(raw tea.Msg) (tea.Cmd, bool) {
	s := a.env.Styles
	switch msg := raw.(type) {
	case tea.ColorProfileMsg:
		if s.profile == msg.Profile {
			return nil, true
		}
		s.setProfile(msg.Profile)
		a.colorGeneration++
		cmds := []tea.Cmd{a.refreshStyles()}
		if s.profile >= colorprofile.ANSI256 {
			if !a.colorsQueried {
				a.colorsQueried = true
				cmds = append(cmds, tea.RequestForegroundColor, tea.RequestBackgroundColor)
			}
			if a.opts.ColorQueryTimeout > 0 && (s.background == nil || s.foreground == nil) {
				generation := a.colorGeneration
				cmds = append(cmds, tea.Tick(a.opts.ColorQueryTimeout, func(time.Time) tea.Msg { return colorQueryTimeoutMsg{generation} }))
			}
		}
		return tea.Batch(cmds...), true
	case tea.ForegroundColorMsg:
		s.foreground = msg.Color
		return nil, true
	case tea.BackgroundColorMsg:
		s.background = msg.Color
		return nil, true
	case colorQueryTimeoutMsg:
		if msg.generation != a.colorGeneration || s.profile < colorprofile.ANSI256 || a.colorNotice {
			return nil, true
		}
		var missing []string
		if s.background == nil {
			missing = append(missing, "background")
		}
		if s.foreground == nil {
			missing = append(missing, "foreground")
		}
		if len(missing) > 0 {
			a.colorNotice = true
			a.log.Add(LevelInfo, "numeric shading unavailable: terminal did not report "+strings.Join(missing, " and "))
		}
		return nil, true
	}
	return nil, false
}
