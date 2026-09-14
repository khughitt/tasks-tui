package ui

import (
	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/launch"
)

// openLaunch is spec §7: pick a harness, spawn it in the task's checkout, touch nothing.
func (a *App) openLaunch() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	names := a.env.Config.Launch.Names()
	if len(names) == 0 {
		a.log.Add(LevelError, "launch: no harness configured")
		return nil
	}
	a.overlay = &pickerOverlay{styles: a.env.Styles, title: "launch on " + tgt.ID, items: names,
		onPick: func(name string) tea.Cmd {
			co := a.env.checkout(tgt)
			if co.Notice != "" {
				a.log.Add(LevelWarning, co.Notice)
			}
			cfg, environ, spawn := a.env.Config.Launch, a.env.Environ, a.env.Spawn
			return func() tea.Msg {
				plan, err := launch.Build(cfg, name, co.Dir, tgt.ID, tgt.Title, environ)
				if err != nil {
					return launchMsg{harness: name, id: tgt.ID, err: err}
				}
				if err := spawn(plan); err != nil {
					return launchMsg{harness: name, id: tgt.ID, err: err}
				}
				return launchMsg{harness: name, id: tgt.ID, dir: co.Dir, noPrompt: plan.NoPrompt}
			}
		}}
	return nil
}
