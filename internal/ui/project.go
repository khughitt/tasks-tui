package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type stubProject struct {
	env    *Env
	prefix string
	ready  int
}
type projectStubMsg struct {
	ready int
	err   error
}

func newProjectView(env *Env, prefix string) view { return &stubProject{env: env, prefix: prefix} }
func (p *stubProject) title() string              { return p.prefix }
func (p *stubProject) reload() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := p.env.ctx()
		defer cancel()
		res, err := p.env.Client.Ready(ctx, p.prefix)
		return projectStubMsg{len(res.Tasks), err}
	}
}
func (p *stubProject) update(msg tea.Msg) (view, tea.Cmd) {
	if m, ok := msg.(projectStubMsg); ok {
		if m.err != nil {
			return p, notices(LevelError, m.err.Error())
		}
		p.ready = m.ready
	}
	return p, nil
}
func (p *stubProject) render(_, _ int) string { return fmt.Sprintf("%d Ready", p.ready) }
func (p *stubProject) current() *target       { return nil }
func (p *stubProject) project() string        { return p.prefix }
func (p *stubProject) capturing() bool        { return false }
