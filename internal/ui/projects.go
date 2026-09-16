package ui

import (
	"fmt"
	"sort"
	"strconv"
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
	order      sortBinding
	filter     string
	filtering  bool
	seq        int
}

func newProjectsView(env *Env) *projectsView {
	return &projectsView{env: env, main: NewLoader(), pane: NewLoader(), order: sortBinding{seq: "S a", column: "activity", desc: true}}
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
func (v *projectsView) loading() bool    { return v.main.InFlight() || v.pane.InFlight() }
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
		a, b := v.rows[i], v.rows[j]
		if v.order.desc {
			a, b = b, a
		}
		return projectLess(v.order.column, a, b)
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

func projectLess(column string, a, b tasksctl.Project) bool {
	if column == "prefix" {
		return a.Prefix < b.Prefix
	}
	return deref(a.LastActivity) < deref(b.LastActivity)
}

// projectsTable is spec v1.1 §5. No flexible column and no gutter: the accent bar is the row's identity.
var projectsTable = table{gap: 2, cols: []column{
	{key: "project", label: "project", min: 7},
	{key: "name", label: "name", min: 4, drop: 1},
	{key: "doing", label: "doing", align: alignRight, min: 5},
	{key: "todo", label: "todo", align: alignRight, min: 4},
	{key: "idea", label: "idea", align: alignRight, min: 4},
	{key: "blocked", label: "blocked", align: alignRight, min: 7},
	{key: "activity", label: "activity", align: alignRight, min: 8},
}}

// count is a number that is muted at zero and hot otherwise.
func count(s *Styles, n int, hot lipgloss.Style) cell {
	if n == 0 {
		return text(s.Muted, "0")
	}
	return text(hot, strconv.Itoa(n))
}

// unreachableLine is spec v1.1 §5: the bar and prefix muted, then the indicator across
// the rest of the row, so it survives the name column dropping.
func (v *projectsView) unreachableLine(p tasksctl.Project, widths []int, tableW int, base lipgloss.Style) string {
	s := v.env.Styles
	line := s.Muted.Inherit(base).Render("▌ " + pad(p.Prefix, widths[0]-2) + "  ✗ unreachable")
	return fit(line, tableW, base)
}

func (v *projectsView) projectCells(p tasksctl.Project, now time.Time) []cell {
	s := v.env.Styles
	acc := s.Accent(v.env.slot(p.Prefix))
	project := cell{spans: []span{{"▌ ", acc}, {p.Prefix, s.Bold}}}
	if !p.Reachable || p.Counts == nil {
		// Measured for the project column only; render draws the row through unreachableLine.
		return []cell{{spans: []span{{"▌ ", s.Muted}, {p.Prefix, s.Muted}}}, text(s.Muted, ""), text(s.Muted, ""), text(s.Muted, ""), text(s.Muted, ""), text(s.Muted, ""), text(s.Muted, "")}
	}
	c := *p.Counts
	age := ""
	if p.LastActivity != nil {
		age = ago(*p.LastActivity, now)
	}
	return []cell{project, text(s.Muted, name(p.Root)), count(s, c.Doing, acc), count(s, c.Todo, s.Base), count(s, c.Idea, s.Base), count(s, c.Blocked, s.Error.UnsetBold()), text(s.ageStyle(age), age)}
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
			if b, ok := findSort(projectsSort, msg.String()); ok {
				v.order = b
				v.sortAndFilter()
			}
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
	s, now := v.env.Styles, time.Now()
	cells := make([][]cell, len(v.rows))
	for i, p := range v.rows {
		cells[i] = v.projectCells(p, now)
	}
	widths := projectsTable.widths(cells, width)
	tableW := projectsTable.total(widths)
	var paneRows []rowView
	if v.prime != nil && v.primeFor == v.project() {
		paneRows, _ = v.paneRows()
	}
	paneCells := make([][]cell, len(paneRows))
	for i, r := range paneRows {
		paneCells[i] = s.taskCells(r, now)
	}
	paneW := width - tableW - 1
	showPane := paneW >= taskTable.minWidth(paneCells)
	if !showPane {
		tableW = width
	}
	lines := []string{fit(projectsTable.header(widths, s), tableW, lipgloss.NewStyle())}
	visible := height - 2
	if v.filtering || v.filter != "" {
		visible--
	}
	visible = max(1, visible)
	start := max(0, v.sel-visible+1)
	end := min(len(v.rows), start+visible)
	for i := start; i < end; i++ {
		base := lipgloss.NewStyle()
		if i == v.sel {
			base = s.Surface(v.env.slot(v.rows[i].Prefix))
		}
		if p := v.rows[i]; !p.Reachable || p.Counts == nil {
			lines = append(lines, v.unreachableLine(p, widths, tableW, base))
			continue
		}
		lines = append(lines, fit(projectsTable.row(widths, cells[i], base), tableW, base))
	}
	if v.filtering || v.filter != "" {
		lines = append(lines, lipgloss.NewStyle().MaxWidth(tableW).Render(s.Muted.Render("/"+v.filter)))
	}
	body := strings.Join(lines, "\n")
	if showPane {
		pane := s.Muted.Render("…")
		if paneRows != nil {
			pane = v.renderPane(paneW)
		}
		bodyH := max(1, height-1)
		sep := strings.TrimSuffix(strings.Repeat(s.Muted.Render("│")+"\n", bodyH), "\n")
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(tableW).Render(body), sep, lipgloss.NewStyle().Width(paneW).Render(pane))
	}
	strip := lipgloss.NewStyle().MaxWidth(width).Render(s.Muted.Render(fmt.Sprintf("%d doing · %d parked across %d projects", len(v.data.doing.Tasks), len(v.data.parked.Tasks), len(v.data.projects.Projects))))
	return lipgloss.NewStyle().Height(max(0, height-1)).MaxHeight(max(0, height-1)).Render(body) + "\n" + strip
}

const paneReadyCap = 8

func (v *projectsView) paneRows() (rows []rowView, nextSteps map[int]string) {
	rows = []rowView{}
	slot := v.env.slot(v.primeFor)
	nextSteps = map[int]string{}
	for _, r := range v.prime.Doing {
		rows = append(rows, fromRow(r, slot))
	}
	for _, r := range v.prime.Parked {
		nextSteps[len(rows)] = r.Park.NextStep
		rows = append(rows, fromParked(r, slot))
	}
	for i, r := range v.prime.Ready {
		if i == paneReadyCap {
			break
		}
		rows = append(rows, fromRow(r, slot))
	}
	return rows, nextSteps
}

func (v *projectsView) renderPane(width int) string {
	s := v.env.Styles
	slot := v.env.slot(v.primeFor)
	rows, nextSteps := v.paneRows()
	if len(rows) == 0 {
		return s.Muted.Render("nothing doing, parked, or ready")
	}
	rt := s.layoutRows(rows, width, time.Now())
	indent := strings.Repeat(" ", rt.titleOffset())
	var out []string
	section := func(title string, n int) {
		if n > 0 {
			out = append(out, s.Accent(slot).Render(title))
		}
	}
	i := 0
	emit := func(n int) {
		for end := i + n; i < end; i++ {
			out = append(out, rt.line(i, false))
			if step, ok := nextSteps[i]; ok {
				out = append(out, lipgloss.NewStyle().MaxWidth(width).Render(indent+s.Muted.Render("→ "+step)))
			}
		}
	}
	section("doing", len(v.prime.Doing))
	emit(len(v.prime.Doing))
	section("parked", len(v.prime.Parked))
	emit(len(v.prime.Parked))
	section("ready", len(v.prime.Ready))
	emit(min(len(v.prime.Ready), paneReadyCap))
	if n := len(v.prime.Ready); n > paneReadyCap {
		out = append(out, s.Muted.Render(fmt.Sprintf("%s%d ready · %d shown", indent, n, paneReadyCap)))
	}
	return strings.Join(out, "\n")
}
