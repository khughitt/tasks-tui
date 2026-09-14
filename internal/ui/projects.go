package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"tasks-tui/internal/tasksctl"
)

const paneDebounce = 150 * time.Millisecond

type projectsData struct {
	projects tasksctl.ProjectsResult
	doing    tasksctl.RowsResult
	parked   tasksctl.ParkedResult
}
type paneData struct {
	prefix string
	res    tasksctl.PrimeResult
}
type paneError struct {
	prefix string
	err    error
}

func (e paneError) Error() string { return e.err.Error() }

type paneDebounceMsg struct{ seq int }
type projectsView struct {
	env        *Env
	main, pane *Loader
	data       projectsData
	prime      *tasksctl.PrimeResult
	primeFor   string
	rows       []tasksctl.Project
	sel        int
	byPrefix   bool
	filter     string
	filtering  bool
	seq        int
}

func newProjectsView(env *Env) *projectsView {
	return &projectsView{env: env, main: NewLoader(), pane: NewLoader()}
}
func (v *projectsView) title() string { return "projects" }
func (v *projectsView) project() string {
	if v.sel >= 0 && v.sel < len(v.rows) && v.rows[v.sel].Reachable {
		return v.rows[v.sel].Prefix
	}
	return ""
}
func (v *projectsView) current() *target { return nil }
func (v *projectsView) capturing() bool  { return v.filtering }
func (v *projectsView) reload() tea.Cmd {
	return v.main.Request(func(gen uint64) tea.Cmd {
		return v.main.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			var d projectsData
			var err error
			if d.projects, err = v.env.Client.Projects(ctx); err != nil {
				return nil, err
			}
			if d.doing, err = v.env.Client.DoingAll(ctx); err != nil {
				return nil, err
			}
			if d.parked, err = v.env.Client.Parked(ctx, ""); err != nil {
				return nil, err
			}
			return d, nil
		})
	})
}
func (v *projectsView) loadPane() tea.Cmd {
	prefix := v.project()
	if prefix == "" {
		return nil
	}
	return v.pane.Request(func(gen uint64) tea.Cmd {
		return v.pane.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			res, err := v.env.Client.Prime(ctx, prefix)
			if err != nil {
				return nil, paneError{prefix, err}
			}
			return paneData{prefix, res}, nil
		})
	})
}
func (v *projectsView) debouncePane() tea.Cmd {
	v.seq++
	seq := v.seq
	return tea.Tick(paneDebounce, func(time.Time) tea.Msg { return paneDebounceMsg{seq} })
}
func (v *projectsView) sortAndFilter() {
	oldPos, selected := v.sel, ""
	if oldPos >= 0 && oldPos < len(v.rows) {
		selected = v.rows[oldPos].Prefix
	}
	v.rows = v.rows[:0]
	for _, p := range v.data.projects.Projects {
		if v.filter == "" || strings.Contains(p.Prefix, v.filter) || strings.Contains(name(p.Root), v.filter) {
			v.rows = append(v.rows, p)
		}
	}
	sort.SliceStable(v.rows, func(i, j int) bool {
		if v.byPrefix {
			return v.rows[i].Prefix < v.rows[j].Prefix
		}
		return deref(v.rows[i].LastActivity) > deref(v.rows[j].LastActivity)
	})
	if len(v.rows) == 0 {
		v.sel = 0
		return
	}
	v.sel = min(oldPos, len(v.rows)-1)
	for i, p := range v.rows {
		if p.Prefix == selected {
			v.sel = i
			break
		}
	}
}
func name(root string) string {
	if i := strings.LastIndex(root, "/"); i >= 0 {
		return root[i+1:]
	}
	return root
}
func (v *projectsView) update(raw tea.Msg) (view, tea.Cmd) {
	switch msg := raw.(type) {
	case loadMsg:
		switch msg.loader {
		case v.main.ID():
			accept, next := v.main.Done(msg.gen)
			if !accept || msg.background {
				return v, next
			}
			if msg.err != nil {
				return v, tea.Batch(next, notices(LevelError, msg.err.Error()))
			}
			v.data = msg.data.(projectsData)
			v.sortAndFilter()
			return v, tea.Batch(next, v.loadPane(), v.warnings())
		case v.pane.ID():
			accept, next := v.pane.Done(msg.gen)
			if !accept || msg.background {
				return v, next
			}
			if msg.err != nil {
				if pe, ok := msg.err.(paneError); ok && pe.prefix != v.project() {
					return v, next
				}
				return v, tea.Batch(next, notices(LevelError, msg.err.Error()))
			}
			d := msg.data.(paneData)
			if d.prefix != v.project() {
				return v, next
			}
			v.prime, v.primeFor = &d.res, d.prefix
			return v, tea.Batch(next, notices(LevelWarning, prefixed("prime --project "+d.prefix, d.res.Warnings)...))
		}
	case paneDebounceMsg:
		if msg.seq == v.seq {
			return v, v.loadPane()
		}
	case tea.KeyPressMsg:
		if v.filtering {
			switch msg.String() {
			case "enter", "esc":
				v.filtering = false
			case "backspace":
				if len(v.filter) > 0 {
					_, n := utf8.DecodeLastRuneInString(v.filter)
					v.filter = v.filter[:len(v.filter)-n]
				}
			default:
				if msg.Text != "" {
					v.filter += msg.Text
				}
			}
			v.sortAndFilter()
			return v, v.debouncePane()
		}
		switch {
		case key.Matches(msg, keys.Down):
			if v.sel < len(v.rows)-1 {
				v.sel++
			}
			return v, v.debouncePane()
		case key.Matches(msg, keys.Up):
			if v.sel > 0 {
				v.sel--
			}
			return v, v.debouncePane()
		case key.Matches(msg, keys.Sort):
			v.byPrefix = !v.byPrefix
			v.sortAndFilter()
			return v, v.debouncePane()
		case key.Matches(msg, keys.Filter):
			v.filtering = true
			return v, nil
		case key.Matches(msg, keys.Enter):
			if p := v.project(); p != "" {
				return v, func() tea.Msg { return pushMsg{newProjectView(v.env, p)} }
			}
		}
	}
	return v, nil
}
func (v *projectsView) warnings() tea.Cmd {
	ws := append([]string{}, prefixed("projects", v.data.projects.Warnings)...)
	ws = append(ws, prefixed("list --all-projects --status doing", v.data.doing.Warnings)...)
	ws = append(ws, prefixed("list --all-projects --parked", v.data.parked.Warnings)...)
	return notices(LevelWarning, ws...)
}
func (v *projectsView) render(width, height int) string {
	s := v.env.Styles
	left := width * 5 / 9
	right := max(0, width-left-1)
	lines := []string{s.Header.Render(pad("project", 14)) + s.Muted.Render(pad("doing", 6)+pad("todo", 6)+pad("idea", 6)+pad("blocked", 8)+"activity")}
	visible := height - 1 - lipgloss.Height(lipgloss.NewStyle().Width(left).Render(lines[0]))
	if v.filtering || v.filter != "" {
		visible--
	}
	visible = max(1, visible)
	start := max(0, v.sel-visible+1)
	end := min(len(v.rows), start+visible)
	for i := start; i < end; i++ {
		p := v.rows[i]
		var line string
		if !p.Reachable || p.Counts == nil {
			line = s.Muted.Render(pad(p.Prefix+" ✗ unreachable", left))
		} else {
			acc := s.Accent(v.env.slot(p.Prefix))
			c := *p.Counts
			line = acc.Render("▌ ") + pad(p.Prefix, 8) + s.Muted.Render(pad(name(p.Root), 4))
			line = pad(line, 14) + pad(fmt.Sprint(c.Doing), 6) + pad(fmt.Sprint(c.Todo), 6) + pad(fmt.Sprint(c.Idea), 6) + pad(fmt.Sprint(c.Blocked), 8)
			if p.LastActivity != nil {
				line += ago(*p.LastActivity, time.Now())
			}
		}
		st := lipgloss.NewStyle().MaxWidth(left)
		if i == v.sel {
			st = st.Reverse(true)
		}
		lines = append(lines, st.Render(pad(line, left)))
	}
	if v.filtering || v.filter != "" {
		lines = append(lines, s.Muted.Render("/"+v.filter))
	}
	table := strings.Join(lines, "\n")
	pane := s.Muted.Render("…")
	if v.prime != nil && v.primeFor == v.project() {
		pane = v.renderPane(right)
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(left).Render(table), " ", lipgloss.NewStyle().Width(right).Render(pane))
	strip := s.Muted.Render(fmt.Sprintf("%d doing · %d parked across %d projects", len(v.data.doing.Tasks), len(v.data.parked.Tasks), len(v.data.projects.Projects)))
	return lipgloss.NewStyle().Height(max(0, height-1)).MaxHeight(max(0, height-1)).Render(body) + "\n" + strip
}
func (v *projectsView) renderPane(width int) string {
	s := v.env.Styles
	slot := v.env.slot(v.primeFor)
	var out []string
	section := func(title string, n int) {
		if n > 0 {
			out = append(out, s.Accent(slot).Render(title))
		}
	}
	section("doing", len(v.prime.Doing))
	for _, r := range v.prime.Doing {
		out = append(out, s.renderRow(fromRow(r, slot), width, false))
	}
	section("parked", len(v.prime.Parked))
	for _, r := range v.prime.Parked {
		out = append(out, s.renderRow(fromParked(r, slot), width, false), s.Muted.Render("    → "+r.Park.NextStep))
	}
	section("ready", len(v.prime.Ready))
	for i, r := range v.prime.Ready {
		if i == 8 {
			out = append(out, s.Muted.Render(fmt.Sprintf("    … %d more", len(v.prime.Ready)-8)))
			break
		}
		out = append(out, s.renderRow(fromRow(r, slot), width, false))
	}
	if len(out) == 0 {
		out = append(out, s.Muted.Render("nothing doing, parked, or ready"))
	}
	return strings.Join(out, "\n")
}
