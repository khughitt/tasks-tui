package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/khughitt/tasks-tui/internal/config"
	"github.com/khughitt/tasks-tui/internal/launch"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

type Env struct {
	Client  *tasksctl.Client
	Styles  *Styles
	Config  config.Config
	Roots   map[string]string
	Exists  func(string) bool
	Environ []string
	Spawn   func(launch.Plan) error
	Timeout time.Duration
}

func (e *Env) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), e.Timeout)
}

// SetProjects seeds startup identity and refreshes it only from accepted CLI data.
func (e *Env) SetProjects(projects []tasksctl.Project) {
	e.Roots = make(map[string]string, len(projects))
	e.Styles.projects = make(map[string]tasksctl.Project, len(projects))
	for _, p := range projects {
		e.Roots[p.Prefix] = p.Root
		e.Styles.projects[p.Prefix] = p
	}
}

type target struct {
	ID, Title, Prefix string
	Park              *tasksctl.ParkInfo
	Claim             *tasksctl.ClaimInfo
	Unresolved        bool
}

func (e *Env) checkout(t target) tasksctl.Checkout {
	return tasksctl.CheckoutFor(t.Park, t.Claim, e.Roots[t.Prefix], e.Exists)
}

type view interface {
	title() string
	reload() tea.Cmd
	update(tea.Msg) (view, tea.Cmd)
	render(width, height int) string
	current() *target
	project() string
	capturing() bool
	loading() bool
	// warnings is the warnings[] of the view's latest accepted load, prefixed with the
	// command: the state of its scope, replaced by every reload (messages and warnings,
	// docs/guide.md).
	warnings() []string
}

type overlay interface {
	update(tea.Msg) (overlay, tea.Cmd)
	render(width int) string
}

type tickMsg struct{}
type noticesMsg []Message

func notices(level Level, texts ...string) tea.Cmd {
	if len(texts) == 0 {
		return nil
	}
	ms := make(noticesMsg, len(texts))
	for i, text := range texts {
		ms[i] = Message{Level: level, Text: text}
	}
	return func() tea.Msg { return ms }
}

func prefixed(cmd string, warnings []string) []string {
	out := make([]string, len(warnings))
	for i, warning := range warnings {
		out[i] = cmd + ": " + warning
	}
	return out
}

type writeMsg struct {
	action string
	tgt    target
	force  bool
	res    tasksctl.WriteResult
	err    error
}
type launchMsg struct {
	harness, id, dir string
	prompt           string
	noPrompt         bool
	err              error
}
type pushMsg struct{ v view }
