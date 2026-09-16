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

type tab int

const (
	tabReady tab = iota
	tabDoing
	tabOpen
	tabIdeas
	tabDone
)

var tabNames = [...]string{"Ready", "Doing", "Open", "Ideas", "Done"}

type projectData struct {
	tab      tab
	rows     []rowView
	counts   tasksctl.Counts
	warnings []string
}

type projectView struct {
	env        *Env
	prefix     string
	loader     *Loader
	tab        tab
	all, rows  []rowView
	counts     tasksctl.Counts
	sel        int
	filter     string
	filtering  bool
	offset     int
	loadedTab  tab
	hasData    bool
	sortKey    string
	descending bool
	shown      [len(tabNames)]int
	hasCount   [len(tabNames)]bool
}

func newProjectView(env *Env, prefix string) *projectView {
	return &projectView{env: env, prefix: prefix, loader: NewLoader()}
}
func (v *projectView) title() string   { return v.prefix }
func (v *projectView) project() string { return v.prefix }
func (v *projectView) capturing() bool { return v.filtering }
func (v *projectView) loading() bool   { return v.loader.InFlight() }
func (v *projectView) current() *target {
	if v.sel < 0 || v.sel >= len(v.rows) || v.rows[v.sel].Unresolved {
		return nil
	}
	t := v.rows[v.sel].target()
	return &t
}

func (v *projectView) reload() tea.Cmd {
	tab, prefix, slot := v.tab, v.prefix, v.env.slot(v.prefix)
	return v.loader.Request(func(gen uint64) tea.Cmd {
		return v.loader.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			prime, err := v.env.Client.Prime(ctx, prefix)
			if err != nil {
				return nil, err
			}
			d := projectData{tab: tab, counts: prime.Counts, warnings: prefixed("prime --project "+prefix, prime.Warnings)}
			add := func(command string, res tasksctl.RowsResult, err error) error {
				if err != nil {
					return err
				}
				for _, r := range res.Tasks {
					d.rows = append(d.rows, fromRow(r, slot))
				}
				d.warnings = append(d.warnings, prefixed(command, res.Warnings)...)
				return nil
			}
			c := v.env.Client
			switch tab {
			case tabReady:
				res, err := c.Ready(ctx, prefix)
				if err := add("ready --project "+prefix, res, err); err != nil {
					return nil, err
				}
			case tabDoing:
				command := "list --project " + prefix + " --status doing"
				res, err := c.List(ctx, prefix, tasksctl.ListOpts{Status: []string{"doing"}})
				if err := add(command, res, err); err != nil {
					return nil, err
				}
				parked, err := c.Parked(ctx, prefix)
				if err != nil {
					return nil, err
				}
				byID := make(map[string]int, len(d.rows))
				for i, r := range d.rows {
					byID[r.ID] = i
				}
				for _, r := range parked.Tasks {
					if i, ok := byID[r.ID]; ok {
						d.rows[i].Park = r.Park
						continue
					}
					d.rows = append(d.rows, fromParked(r, slot))
				}
				d.warnings = append(d.warnings, prefixed("list --project "+prefix+" --parked", parked.Warnings)...)
			case tabOpen:
				res, err := c.List(ctx, prefix, tasksctl.ListOpts{})
				if err := add("list --project "+prefix, res, err); err != nil {
					return nil, err
				}
			case tabIdeas:
				command := "list --project " + prefix + " --status idea"
				res, err := c.List(ctx, prefix, tasksctl.ListOpts{Status: []string{"idea"}})
				if err := add(command, res, err); err != nil {
					return nil, err
				}
			case tabDone:
				command := "list --project " + prefix + " --status done --sort updated"
				res, err := c.List(ctx, prefix, tasksctl.ListOpts{Status: []string{"done"}, SortUpdated: true})
				if err := add(command, res, err); err != nil {
					return nil, err
				}
			}
			return d, nil
		})
	})
}

func (v *projectView) applyFilter() {
	oldPos, selected := v.sel, ""
	if oldPos >= 0 && oldPos < len(v.rows) {
		selected = v.rows[oldPos].ID
	}
	v.rows = v.rows[:0]
	needle := strings.ToLower(v.filter)
	for _, r := range v.all {
		if needle == "" || strings.Contains(strings.ToLower(r.ID), needle) || strings.Contains(strings.ToLower(r.Title), needle) || strings.Contains(strings.ToLower(strings.Join(r.Tags, " ")), needle) {
			v.rows = append(v.rows, r)
		}
	}
	if v.sortKey != "" {
		sort.SliceStable(v.rows, func(i, j int) bool { return v.less(v.rows[i], v.rows[j]) })
	}
	if v.hasData {
		v.shown[v.loadedTab], v.hasCount[v.loadedTab] = len(v.rows), true
	}
	if len(v.rows) == 0 {
		v.sel = 0
		return
	}
	v.sel = min(oldPos, len(v.rows)-1)
	for i, r := range v.rows {
		if r.ID == selected {
			v.sel = i
			break
		}
	}
}

func (v *projectView) cycleSort(key string) {
	switch {
	case v.sortKey != key:
		v.sortKey, v.descending = key, false
	case !v.descending:
		v.descending = true
	default:
		v.sortKey, v.descending = "", false
	}
	v.applyFilter()
}

func (v *projectView) less(a, b rowView) bool {
	missing := func(r rowView) bool {
		switch v.sortKey {
		case "prio":
			return r.Priority == nil
		case "age":
			_, err := time.Parse(time.RFC3339, r.Updated)
			return err != nil
		case "status":
			return r.Unresolved || r.Status == ""
		default:
			return v.value(r) == ""
		}
	}
	am, bm := missing(a), missing(b)
	if am && bm {
		return strings.Compare(a.ID, b.ID) < 0
	}
	if am != bm {
		return !am
	}
	var order int
	switch v.sortKey {
	case "prio":
		order = *a.Priority - *b.Priority
	case "age":
		at, _ := time.Parse(time.RFC3339, a.Updated)
		bt, _ := time.Parse(time.RFC3339, b.Updated)
		if at.Before(bt) {
			order = 1
		} else if at.After(bt) {
			order = -1
		}
	default:
		order = strings.Compare(v.value(a), v.value(b))
	}
	if order == 0 {
		order = strings.Compare(a.ID, b.ID)
	}
	if v.descending {
		order = -order
	}
	return order < 0
}

func (v *projectView) value(r rowView) string {
	switch v.sortKey {
	case "id":
		return r.ID
	case "size":
		return rank(r.Size, map[string]int{"xs": 0, "s": 1, "m": 2, "l": 3, "xl": 4})
	case "cx":
		return rank(r.Complexity, map[string]int{"low": 0, "mid": 1, "high": 2})
	case "status":
		// Keep this rank aligned with tasks' Status enum.
		return rank(r.Status, map[string]int{"idea": 0, "todo": 1, "doing": 2, "blocked": 3, "shelved": 4, "done": 5, "dropped": 6})
	case "proc":
		return r.Process
	case "title":
		return strings.ToLower(r.Title)
	}
	return ""
}

func rank(value string, ranks map[string]int) string {
	n, ok := ranks[value]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func (v *projectView) setTab(t tab) tea.Cmd { v.tab, v.sel, v.offset = t, 0, 0; return v.reload() }

func (v *projectView) update(msg tea.Msg) (view, tea.Cmd) {
	switch msg := msg.(type) {
	case loadMsg:
		if msg.loader != v.loader.ID() {
			return v, nil
		}
		accept, next := v.loader.Done(msg.gen)
		if !accept || msg.background {
			return v, next
		}
		if msg.err != nil {
			return v, tea.Batch(next, notices(LevelError, msg.err.Error()))
		}
		d := msg.data.(projectData)
		if d.tab == v.tab {
			if v.loadedTab != d.tab {
				v.rows, v.sel = nil, 0
			}
			v.all, v.counts = d.rows, d.counts
			v.loadedTab, v.hasData = d.tab, true
			v.applyFilter()
		}
		return v, tea.Batch(next, notices(LevelWarning, d.warnings...))
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
			v.applyFilter()
			return v, nil
		}
		switch {
		case key.Matches(msg, keys.Down):
			v.sel = min(v.sel+1, max(0, len(v.rows)-1))
		case key.Matches(msg, keys.Up):
			v.sel = max(0, v.sel-1)
		case key.Matches(msg, keys.Top):
			v.sel = 0
		case key.Matches(msg, keys.Bottom):
			v.sel = max(0, len(v.rows)-1)
		case key.Matches(msg, keys.Tab, keys.Next):
			return v, v.setTab((v.tab + 1) % tab(len(tabNames)))
		case key.Matches(msg, keys.ShiftTab, keys.Prev):
			return v, v.setTab((v.tab + tab(len(tabNames)) - 1) % tab(len(tabNames)))
		case key.Matches(msg, keys.Filter):
			v.filtering = true
		case key.Matches(msg, keys.Enter):
			if t := v.current(); t != nil {
				return v, func() tea.Msg { return pushMsg{v: newTaskView(v.env, *t)} }
			}
		case key.Matches(msg, keys.Sort):
			if b, ok := findSort(tasksSort, msg.String()); ok {
				v.sortKey, v.descending = b.column, b.desc
				if v.sortKey == "priority" {
					v.sortKey = "prio"
				}
				v.applyFilter()
			}
		case key.Matches(msg, keys.Digits):
			return v, v.setTab(tabFor(msg.String()))
		}
	}
	return v, nil
}

var emptyText = map[tab]string{
	tabReady: "nothing ready — a adds a task",
	tabDoing: "nothing doing",
	tabOpen:  "no open tasks",
	tabIdeas: "no ideas — a then ? files one",
	tabDone:  "nothing done yet",
}

func (v *projectView) chip(n int, label string, hot lipgloss.Style) string {
	num := v.env.Styles.Muted.Render(fmt.Sprint(n))
	if n > 0 {
		num = hot.Render(fmt.Sprint(n))
	}
	return num + " " + v.env.Styles.Muted.Render(label)
}

func (v *projectView) render(width, height int) string {
	s, root, slot := v.env.Styles, v.env.Roots[v.prefix], v.env.slot(v.prefix)
	c := v.counts
	head := s.Accent(slot).Render("▌ ") + s.Header.Render(v.prefix) + "  " + s.Muted.Render(name(root)+"  "+root) + "   " + strings.Join([]string{v.chip(c.Doing, "doing", s.Accent(slot)), v.chip(c.Todo, "todo", s.Base), v.chip(c.Idea, "idea", s.Base), v.chip(c.Blocked, "blocked", s.Error.UnsetBold())}, "  ")
	tabs := make([]string, len(tabNames))
	for i, n := range tabNames {
		label := fmt.Sprintf("%d %s", i+1, n)
		if tab(i) != v.tab {
			tabs[i] = s.Muted.Render(" " + label + " ")
			continue
		}
		if v.hasCount[i] {
			label += " " + fmt.Sprint(v.shown[i])
		}
		tabs[i] = s.Pill(slot).Render(" " + label + " ")
	}
	rt := s.layoutRowsWithLabels(v.rows, width, time.Now(), v.headerLabels())
	clipped := lipgloss.NewStyle().MaxWidth(width)
	lines := []string{clipped.Render(head), clipped.Render(strings.Join(tabs, " ")), rt.header()}
	filtering := v.filtering || v.filter != ""
	avail := height - len(lines)
	if filtering {
		avail--
	}
	avail = max(1, avail)
	if v.sel < v.offset {
		v.offset = v.sel
	}
	if v.sel >= v.offset+avail {
		v.offset = v.sel - avail + 1
	}
	for i := v.offset; i < len(v.rows) && i < v.offset+avail; i++ {
		lines = append(lines, rt.line(i, i == v.sel))
	}
	if len(v.rows) == 0 && v.hasData && v.loadedTab == v.tab {
		if v.filter != "" {
			lines = append(lines, "  "+s.Muted.Render("no rows match /"+v.filter))
		} else {
			lines = append(lines, "  "+s.Muted.Render(emptyText[v.tab]))
		}
	}
	if filtering {
		lines = append(lines, s.Muted.Render(fmt.Sprintf("/%s · %d of %d", v.filter, len(v.rows), len(v.all))))
	}
	return lipgloss.NewStyle().MaxHeight(height).Render(strings.Join(lines, "\n"))
}

func (v *projectView) headerLabels() map[string]string {
	if v.sortKey == "" {
		return nil
	}
	key := map[string]string{"prio": "prio", "size": "size", "cx": "cx", "proc": "proc", "age": "age"}[v.sortKey]
	if key == "" {
		key = v.sortKey
	}
	arrow := "↑"
	if v.descending {
		arrow = "↓"
	}
	for _, c := range taskTable.cols {
		if c.key == key {
			return map[string]string{key: c.label + arrow}
		}
	}
	return nil
}

// tabFor is the select dispatch: digit key "1".."5" to its tab. The conformance test
// derives the "tab N" argument from it.
func tabFor(k string) tab { return tab(k[0] - '1') }
