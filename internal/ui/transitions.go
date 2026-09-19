package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

var parkReasons = []string{"", "review", "decision", "approval", "environment", "dependency", "session"}

func (a *App) selected() *target {
	t := a.top().current()
	if t == nil {
		a.log.Add(LevelInfo, "no task highlighted")
	}
	return t
}

func (a *App) write(action string, tgt target, force bool, run func(dir string) (tasksctl.WriteResult, error)) tea.Cmd {
	co := a.env.checkout(tgt)
	if co.Notice != "" {
		a.log.Add(LevelWarning, co.Notice)
	}
	return func() tea.Msg {
		res, err := run(co.Dir)
		return writeMsg{action: action, tgt: tgt, force: force, res: res, err: err}
	}
}

func (a *App) startTarget(force bool) tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	return a.startWith(*t, force)
}

func (a *App) startWith(tgt target, force bool) tea.Cmd {
	return a.write("start", tgt, force, func(dir string) (tasksctl.WriteResult, error) {
		ctx, cancel := a.env.ctx()
		defer cancel()
		return a.env.Client.Start(ctx, dir, tgt.ID, force)
	})
}

func (a *App) openForce(tgt target, detail string) tea.Cmd {
	a.overlay = &confirmOverlay{styles: a.env.Styles, text: a.env.Styles.Warning.Render(detail), accept: "F", label: "start --force",
		onAccept: func() tea.Cmd { return a.startWith(tgt, true) }}
	return nil
}

func (a *App) openPark() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	s := a.env.Styles
	spec := tasksctl.ParkSpec{}
	reason := 0
	p := newPrompt(s, "park "+tgt.ID+" — next step", "the one concrete next step")
	p.required = "next step is required"
	p.keys = func(msg tea.KeyPressMsg) bool {
		switch msg.String() {
		case "ctrl+u":
			spec.WaitingOnUser = !spec.WaitingOnUser
			return true
		case "ctrl+r":
			reason = (reason + 1) % len(parkReasons)
			spec.Reason = parkReasons[reason]
			return true
		}
		return false
	}
	p.hint = func(string) string {
		who := "agent"
		if spec.WaitingOnUser {
			who = "user"
		}
		r := spec.Reason
		if r == "" {
			r = "none"
		}
		return s.Muted.Render("ctrl+u ") + "waiting on: " + who + s.Muted.Render("   ctrl+r ") + "reason: " + r
	}
	p.submit = func(value string) (overlay, tea.Cmd) {
		spec.NextStep = strings.TrimSpace(value)
		return nil, a.write("park", tgt, false, func(dir string) (tasksctl.WriteResult, error) {
			ctx, cancel := a.env.ctx()
			defer cancel()
			return a.env.Client.Park(ctx, dir, tgt.ID, spec)
		})
	}
	a.overlay = p
	return nil
}

func (a *App) openDone() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	p := newPrompt(a.env.Styles, "done "+tgt.ID+" — what landed", "optional message")
	p.submit = func(value string) (overlay, tea.Cmd) {
		message := strings.TrimSpace(value)
		return nil, a.write("done", tgt, false, func(dir string) (tasksctl.WriteResult, error) {
			ctx, cancel := a.env.ctx()
			defer cancel()
			return a.env.Client.Done(ctx, dir, tgt.ID, message)
		})
	}
	a.overlay = p
	return nil
}

func (a *App) openDrop() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	s := a.env.Styles
	p := newPrompt(s, "drop "+tgt.ID+" — why", "optional message")
	p.submit = func(value string) (overlay, tea.Cmd) {
		message := strings.TrimSpace(value)
		return &confirmOverlay{styles: s, text: "really drop " + tgt.ID + "? " + s.Muted.Render(strings.TrimSpace(tgt.Title)), accept: "y", label: "drop",
			onAccept: func() tea.Cmd {
				return a.write("drop", tgt, false, func(dir string) (tasksctl.WriteResult, error) {
					ctx, cancel := a.env.ctx()
					defer cancel()
					return a.env.Client.Drop(ctx, dir, tgt.ID, message)
				})
			}}, nil
	}
	a.overlay = p
	return nil
}
