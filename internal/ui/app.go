package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tasks-tui/internal/tasksctl"
)

type Options struct {
	Refresh time.Duration
	Stack   []view
}

type App struct {
	env                   *Env
	opts                  Options
	stack                 []view
	overlay               overlay
	log                   Log
	legend, showLog       bool
	logOff, width, height int
}

func New(env *Env, opts Options) *App { return &App{env: env, opts: opts, stack: opts.Stack} }
func (a *App) top() view              { return a.stack[len(a.stack)-1] }
func (a *App) Init() tea.Cmd          { return tea.Batch(a.top().reload(), a.tick()) }
func (a *App) tick() tea.Cmd {
	if a.opts.Refresh <= 0 {
		return nil
	}
	return tea.Tick(a.opts.Refresh, func(time.Time) tea.Msg { return tickMsg{} })
}
func (a *App) push(v view) tea.Cmd { a.stack = append(a.stack, v); return v.reload() }
func (a *App) pop() tea.Cmd {
	if len(a.stack) == 1 {
		return nil
	}
	a.stack = a.stack[:len(a.stack)-1]
	return a.top().reload()
}

func (a *App) Update(raw tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := raw.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		return a, nil
	case tickMsg:
		return a, tea.Batch(a.top().reload(), a.tick())
	case pushMsg:
		return a, a.push(msg.v)
	case loadMsg:
		cmds := make([]tea.Cmd, 0, len(a.stack))
		for i, v := range a.stack {
			msg.background = i != len(a.stack)-1
			nv, cmd := v.update(msg)
			a.stack[i] = nv
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	case noticesMsg:
		for _, notice := range msg {
			a.log.Add(notice.Level, notice.Text)
		}
		return a, nil
	case writeMsg:
		return a, a.afterWrite(msg)
	case addMsg:
		if msg.err != nil {
			a.log.Add(LevelError, "add: "+msg.err.Error())
			return a, nil
		}
		a.log.AddWarnings("add", msg.res.Warnings)
		a.log.Add(LevelInfo, msg.res.Action+" "+msg.res.ID)
		return a, a.top().reload()
	case launchMsg:
		if msg.err != nil {
			a.log.Add(LevelError, "launch: "+msg.err.Error())
			return a, nil
		}
		text := fmt.Sprintf("launched %s on %s in %s", msg.harness, msg.id, msg.dir)
		if msg.noPrompt {
			text += " (no initial prompt)"
		}
		a.log.Add(LevelInfo, text)
		return a, nil
	case tea.PasteMsg:
		if a.overlay != nil {
			var cmd tea.Cmd
			a.overlay, cmd = a.overlay.update(msg)
			return a, cmd
		}
		return a, a.toTop(msg)
	case tea.KeyPressMsg:
		a.log.Dismiss()
		if a.overlay != nil {
			var cmd tea.Cmd
			a.overlay, cmd = a.overlay.update(msg)
			return a, cmd
		}
		if a.top().capturing() {
			return a, a.toTop(msg)
		}
		return a, a.key(msg)
	}
	return a, a.toTop(raw)
}

func (a *App) toTop(msg tea.Msg) tea.Cmd {
	v, cmd := a.top().update(msg)
	a.stack[len(a.stack)-1] = v
	return cmd
}
func (a *App) Notice(level Level, text string) { a.log.Add(level, text) }
func (a *App) Messages() []Message             { return a.log.Entries }

func (a *App) key(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keys.Quit):
		return tea.Quit
	case a.legend || a.showLog:
		if key.Matches(msg, keys.Back, keys.Help, keys.Log) {
			a.legend, a.showLog = false, false
			return nil
		}
		if a.showLog {
			if key.Matches(msg, keys.Up) && a.logOff > 0 {
				a.logOff--
			}
			if key.Matches(msg, keys.Down) {
				a.logOff++
			}
		}
		return nil
	case key.Matches(msg, keys.Help):
		a.legend = true
	case key.Matches(msg, keys.Log):
		a.showLog = true
		a.logOff = max(0, len(a.log.Entries)-(a.height-4))
	case key.Matches(msg, keys.Reload):
		return a.top().reload()
	case key.Matches(msg, keys.Back):
		return a.pop()
	case key.Matches(msg, keys.Add):
		return a.openQuickAdd()
	case key.Matches(msg, keys.Launch):
		return a.openLaunch()
	case key.Matches(msg, keys.Start):
		return a.startTarget(false)
	case key.Matches(msg, keys.Park):
		return a.openPark()
	case key.Matches(msg, keys.Done):
		return a.openDone()
	case key.Matches(msg, keys.Drop):
		return a.openDrop()
	default:
		return a.toTop(msg)
	}
	return nil
}

func (a *App) afterWrite(msg writeMsg) tea.Cmd {
	if msg.err != nil {
		var taskErr *tasksctl.Error
		if errors.As(msg.err, &taskErr) && taskErr.Kind == "claimed" && msg.action == "start" && !msg.force {
			return a.openForce(msg.tgt, taskErr.Detail)
		}
		a.log.Add(LevelError, msg.action+" "+msg.tgt.ID+": "+msg.err.Error())
		return nil
	}
	a.log.AddWarnings(msg.action+" "+msg.tgt.ID, msg.res.Warnings)
	a.log.Add(LevelInfo, msg.action+" "+msg.tgt.ID)
	return a.top().reload()
}

func (a *App) View() tea.View { v := tea.NewView(a.render()); v.AltScreen = true; return v }
func (a *App) render() string {
	if a.width == 0 {
		return ""
	}
	status := a.statusLine()
	bottom := status
	if a.overlay != nil {
		bottom = a.overlay.render(a.width) + "\n" + status
	}
	bodyHeight := a.height - lipgloss.Height(bottom)
	var body string
	switch {
	case a.legend:
		body = a.env.Styles.Header.Render("keys") + "\n" + legendText
	case a.showLog:
		body = a.renderLog(bodyHeight)
	default:
		body = a.top().render(a.width, bodyHeight)
	}
	body = lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(body)
	return body + "\n" + bottom
}
func (a *App) statusLine() string {
	if m := a.log.Current(); m != nil {
		style := a.env.Styles.Info
		if m.Level == LevelWarning {
			style = a.env.Styles.Warning
		} else if m.Level == LevelError {
			style = a.env.Styles.Error
		}
		return style.MaxWidth(a.width).Render(m.Text)
	}
	crumbs := make([]string, len(a.stack))
	for i, v := range a.stack {
		crumbs[i] = v.title()
	}
	return a.env.Styles.Muted.MaxWidth(a.width).Render(strings.Join(crumbs, " › ") + "   ? keys")
}
func (a *App) renderLog(height int) string {
	lines := []string{a.env.Styles.Header.Render("messages") + a.env.Styles.Muted.Render("  (newest last; j/k scroll; esc close)")}
	if a.logOff > len(a.log.Entries) {
		a.logOff = len(a.log.Entries)
	}
	for _, m := range a.log.Entries[a.logOff:] {
		style := a.env.Styles.Info
		if m.Level == LevelWarning {
			style = a.env.Styles.Warning
		} else if m.Level == LevelError {
			style = a.env.Styles.Error
		}
		lines = append(lines, style.MaxWidth(a.width).Render(m.Text))
		if len(lines) >= height {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func (a *App) openLaunch() tea.Cmd { return nil }
