package ui

import tea "charm.land/bubbletea/v2"

func newTaskView(_ *Env, t target) view { return &stubTask{t: t} }

type stubTask struct{ t target }

func (s *stubTask) title() string                  { return s.t.ID }
func (s *stubTask) reload() tea.Cmd                { return nil }
func (s *stubTask) update(tea.Msg) (view, tea.Cmd) { return s, nil }
func (s *stubTask) render(_, _ int) string         { return s.t.ID }
func (s *stubTask) current() *target               { return &s.t }
func (s *stubTask) project() string                { return s.t.Prefix }
func (s *stubTask) capturing() bool                { return false }
