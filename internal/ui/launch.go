package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/khughitt/tasks-tui/internal/launch"
)

// launch resolves a completed c-chord: a fixed harness spawns directly when it is
// configured, c c opens the picker over every configured harness.
func (a *App) launch(seq string) tea.Cmd {
	for _, b := range launchKeys {
		if b.seq != seq {
			continue
		}
		if b.harness == "" {
			return a.openLaunch()
		}
		t := a.selected()
		if t == nil {
			return nil
		}
		if _, ok := a.env.Config.Launch.Harness[b.harness]; !ok {
			a.log.Add(LevelError, "launch: "+b.harness+" is not configured")
			return nil
		}
		return a.launchOn(*t, b.harness)
	}
	return nil
}

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
	a.overlay = &pickerOverlay{styles: a.env.Styles, slot: a.env.slot(tgt.Prefix), title: "launch on " + tgt.ID, items: names,
		onPick: func(name string) tea.Cmd { return a.launchOn(tgt, name) }}
	return nil
}

func (a *App) launchOn(tgt target, name string) tea.Cmd {
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
		return launchMsg{harness: name, id: tgt.ID, dir: co.Dir, prompt: plan.Prompt, noPrompt: plan.NoPrompt}
	}
}
