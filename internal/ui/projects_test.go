package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/uitest"
)

func projectsJSON() string {
	return `{"projects":[{"prefix":"tui","root":"/r/tui","reachable":true,"counts":{"idea":1,"todo":2,"doing":1,"blocked":0,"shelved":0,"done":3,"dropped":0},"total":7,"last_activity":"2026-09-13T10:00:00Z"},{"prefix":"ops","root":"/r/ops","reachable":false,"counts":null,"total":null,"last_activity":null}],"warnings":["registry: ops unreachable"]}`
}
func primeJSON(prefix string) string {
	return `{"prefix":"` + prefix + `","projects":["` + prefix + `"],"counts":{"idea":1,"todo":2,"doing":1,"blocked":0,"shelved":0,"done":3,"dropped":0},"periodic":{"scheduled":0,"next_due":null},"ready":[` + rowJSON(row(prefix+"-aaa111", "todo", 2)) + `],"parked":[],"doing":[` + rowJSON(row(prefix+"-bbb222", "doing", 1)) + `],"roadmap":[],"closeout":[],"warnings":["` + prefix + ` has a stale claim file"]}`
}
func logged(app *App, level Level, text string) bool {
	for _, msg := range app.Messages() {
		if msg.Level == level && strings.Contains(msg.Text, text) {
			return true
		}
	}
	return false
}

func TestProjectsViewLoadsPaneAndStripAndOpens(t *testing.T) {
	f := newFake()
	f.on("", "projects", projectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON(row("tui-bbb222", "doing", 1)))
	f.on("", "list --all-projects --parked", parkedJSON(parkedRow("ops-ccc333", "p", "todo", "/r/ops")))
	f.on("", "prime --project tui", primeJSON("tui"))
	f.on("", "ready --project tui", rowsJSON(row("tui-aaa111", "todo", 2)))
	env := testEnv(f)
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	d.Expect("tui", "ops", "✗ unreachable", "1 doing · 1 parked across 2 projects", "tui-aaa111", "tui-bbb222")
	if got := pv.warnings(); !slices.Equal(got, []string{"projects: registry: ops unreachable", "prime --project tui: tui has a stale claim file"}) {
		t.Fatalf("registry, all-projects, then pane warnings are the view's: %q", got)
	}
	if len(app.Messages()) != 0 {
		t.Fatalf("load warnings stay out of the log: %+v", app.Messages())
	}
	d.Expect("⚠ 2")
	d.Key("j")
	if got := pv.warnings(); len(got) != 1 || pv.project() == "tui" {
		t.Fatalf("a pane loaded for another project lends no warnings: %q (selected %s)", got, pv.project())
	}
	d.Key("k")
	chord(t, d, app, "s", "p")
	d.Key("k")
	d.Key("j")
	d.Expect("tui-aaa111")
	d.Key("enter")
	d.Expect("1 Ready")
	if !f.called("ready --project tui") {
		t.Fatal("enter must open Project view")
	}
	d.Key("esc")
	d.Expect("1 doing · 1 parked across 2 projects")
	d.Key("/")
	d.Type("s")
	d.Key("enter")
	d.ExpectNot("tui-aaa111")
	if len(f.writes()) != 0 {
		t.Fatalf("filtering ran transition: %v", f.writes())
	}
	if f.called("prime --project ops") {
		t.Fatal("unreachable project gets no pane")
	}
	if pv.project() != "" {
		t.Fatalf("unreachable highlight has project %q", pv.project())
	}
	d.Key("enter")
	if len(app.stack) != 1 {
		t.Fatal("enter on unreachable project must be a no-op")
	}
	d.Expect("1 doing · 1 parked across 2 projects")
}

func TestProjectsPaneResultForAnotherPrefixIsDropped(t *testing.T) {
	f := newFake()
	f.on("", "projects", `{"projects":[{"prefix":"tui","root":"/r/tui","reachable":true,"counts":{"idea":1,"todo":2,"doing":1,"blocked":0,"shelved":0,"done":3,"dropped":0},"total":7,"last_activity":"2026-09-13T10:00:00Z"},{"prefix":"ops","root":"/r/ops","reachable":true,"counts":{"idea":0,"todo":1,"doing":0,"blocked":0,"shelved":0,"done":0,"dropped":0},"total":1,"last_activity":"2026-09-12T10:00:00Z"}],"warnings":[]}`)
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project tui", primeJSON("tui"))
	env := testEnv(f)
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	d.Expect("tui-aaa111")
	gen := pv.pane.gen
	d.Feed(loadMsg{loader: pv.pane.ID(), gen: gen, err: paneError{prefix: "ops", err: errors.New("no reply for prime --project ops")}})
	if logged(app, LevelError, "prime --project ops") {
		t.Fatalf("stale pane error logged: %+v", app.Messages())
	}
	d.Feed(loadMsg{loader: pv.pane.ID(), gen: gen, data: paneData{prefix: "ops", res: tasksctl.PrimeResult{Prefix: "ops", Warnings: []string{"w"}}}})
	if slices.Contains(pv.warnings(), "prime --project ops: w") || pv.primeFor != "tui" {
		t.Fatalf("stale pane accepted: %q", pv.warnings())
	}
}

func TestProjectsViewLoadErrorIsShownOnlyWhenCurrent(t *testing.T) {
	f := newFake()
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectsView(env)}})
	d := drive(t, app)
	d.Expect("fake: no reply for projects")
	if len(app.Messages()) != 1 {
		t.Fatalf("got %+v", app.Messages())
	}
}

func TestProjectsSelectionWindowUTF8BackspaceAndReloadPosition(t *testing.T) {
	var entries []string
	for i := range 8 {
		entries = append(entries, fmt.Sprintf(`{"prefix":"p%d","root":"/r/p%d","reachable":false,"counts":null,"total":null,"last_activity":null}`, i, i))
	}
	f := newFake()
	f.on("", "projects", `{"projects":[`+strings.Join(entries, ",")+`],"warnings":[]}`)
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	env := testEnv(f)
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	d.Feed(tea.WindowSizeMsg{Width: 80, Height: 7})
	for range 6 {
		d.Key("j")
	}
	d.Expect("p6")
	d.ExpectNot("p0")
	d.Key("/")
	d.Type("é")
	d.Key("backspace")
	if pv.filter != "" {
		t.Fatalf("UTF-8 backspace left %q", pv.filter)
	}
	pv.sel = 5
	pv.data.projects.Projects = pv.data.projects.Projects[:3]
	pv.sortAndFilter()
	if pv.sel != 2 {
		t.Fatalf("selection = %d, want clamp to 2", pv.sel)
	}
}

func wideProjectsJSON() string {
	return `{"projects":[` +
		`{"prefix":"ai","root":"/r/ai","reachable":true,"counts":{"idea":4,"todo":5,"doing":7,"blocked":0,"shelved":0,"done":0,"dropped":0},"total":16,"last_activity":"2026-09-13T10:00:00Z"},` +
		`{"prefix":"naturalsystemsv2xx","root":"/r/natural-systems-v2","reachable":true,"counts":{"idea":3,"todo":38,"doing":0,"blocked":2,"shelved":0,"done":0,"dropped":0},"total":43,"last_activity":"2026-09-12T10:00:00Z"}` +
		`],"warnings":[]}`
}

// The wide fixture's table is 78 cells (name "natural-systems-v2" is 18), so the pane
// needs a 160-cell terminal: 160-78-1 = 81 >= the task table's 57 for "ai-" ids.
func driveWide(t *testing.T, app *App) *uitest.Driver {
	t.Helper()
	return uitest.New(t, app, 160, 40)
}

func lineWith(t *testing.T, screen, needle string) string {
	t.Helper()
	for _, line := range strings.Split(screen, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	t.Fatalf("no line with %q:\n%s", needle, screen)
	return ""
}

func TestProjectsCountsRightAlignUnderHeadersWhateverThePrefixLength(t *testing.T) {
	f := newFake()
	f.on("", "projects", wideProjectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project ai", primeJSON("ai"))
	f.on("", "prime --project naturalsystemsv2xx", primeJSON("naturalsystemsv2xx"))
	env := testEnv(f)
	d := driveWide(t, New(env, Options{Stack: []view{newProjectsView(env)}}))
	screen := d.Screen()
	head := lineWith(t, screen, "blocked")
	ai := lineWith(t, screen, "▌ ai ")
	ns := lineWith(t, screen, "naturalsystemsv2xx")
	// Cell offsets, not byte offsets: the rows start with a multi-byte accent bar and the header does not.
	doingEnd := col(t, head, "doing") + len("doing")
	if col(t, ai, "7")+1 != doingEnd || col(t, ns, "0")+1 != doingEnd {
		t.Fatalf("doing column right edge %d:\n%s\n%s\n%s", doingEnd, head, ai, ns)
	}
	todoEnd := col(t, head, "todo") + len("todo")
	if col(t, ai, "5")+1 != todoEnd || col(t, ns, "38")+2 != todoEnd {
		t.Fatalf("todo column:\n%s\n%s\n%s", head, ai, ns)
	}
	if !strings.Contains(screen, "│") {
		t.Fatal("a separator stands between the table and the pane")
	}
}

func TestProjectsZeroCountsAreMutedAndBlockedIsRed(t *testing.T) {
	f := newFake()
	f.on("", "projects", wideProjectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project ai", primeJSON("ai"))
	env := testEnv(f)
	pv := newProjectsView(env)
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	_ = d
	s := env.Styles
	cells := pv.projectCells(pv.rows[1], time.Now())
	if got := cells[2].spans[0]; got.text != "0" || got.style.Render("0") != s.Muted.Render("0") {
		t.Fatalf("zero doing muted: %+v", got)
	}
	if got := cells[5].spans[0]; got.text != "2" || got.style.Render("2") != s.Error.UnsetBold().Render("2") {
		t.Fatalf("non-zero blocked red: %+v", got)
	}
}

func TestProjectsPaneHidesUnderTaskMinWidthAndTableClipsNarrower(t *testing.T) {
	f := newFake()
	f.on("", "projects", wideProjectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project ai", primeJSON("ai"))
	env := testEnv(f)
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := driveWide(t, app)
	d.Expect("ai-aaa111", "│")
	d.Feed(tea.WindowSizeMsg{Width: 100, Height: 20})
	d.ExpectNot("ai-aaa111")
	d.ExpectNot("│")
	d.Expect("naturalsystemsv2xx", "38")
	d.Feed(tea.WindowSizeMsg{Width: 48, Height: 20})
	d.Expect("naturalsystemsv2xx")
	d.ExpectNot("natural-systems-v2")
	pv.data.projects.Projects = append(pv.data.projects.Projects, tasksctl.Project{Prefix: "gone", Root: "/r/gone"})
	pv.sortAndFilter()
	d.Expect("gone", "✗ unreachable")
	d.Feed(tea.WindowSizeMsg{Width: 30, Height: 20})
	for _, line := range strings.Split(d.Screen(), "\n") {
		if lipgloss.Width(line) > 30 {
			t.Fatalf("line wider than the terminal: %q", line)
		}
	}
}

func TestProjectsBackgroundResultsAreDrainedAndDropped(t *testing.T) {
	env := testEnv(newFake())
	pv := newProjectsView(env)
	pending := func(gen uint64) tea.Cmd { return func() tea.Msg { return gen } }

	pv.main.Request(pending)
	mainGen := pv.main.gen
	pv.main.Request(pending)
	_, cmd := pv.update(loadMsg{loader: pv.main.ID(), gen: mainGen, data: projectsData{
		projects: tasksctl.ProjectsResult{Projects: []tasksctl.Project{{Prefix: "leak"}}},
	}, err: errors.New("main leak"), background: true})
	if cmd == nil || cmd() != uint64(mainGen+1) || len(pv.data.projects.Projects) != 0 {
		t.Fatal("background main result did not drain pending load or leaked data/notices")
	}

	pv.pane.Request(pending)
	paneGen := pv.pane.gen
	pv.pane.Request(pending)
	_, cmd = pv.update(loadMsg{loader: pv.pane.ID(), gen: paneGen, data: paneData{
		prefix: "tui", res: tasksctl.PrimeResult{Prefix: "leak", Warnings: []string{"leak"}},
	}, err: paneError{prefix: "tui", err: errors.New("pane leak")}, background: true})
	if cmd == nil || cmd() != uint64(paneGen+1) || pv.prime != nil || pv.primeFor != "" {
		t.Fatal("background pane result did not drain pending load or leaked data/notices")
	}
}

func TestProjectsSortChordsUseColumnAndDirection(t *testing.T) {
	env := testEnv(newFake())
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	newer, older := "2026-09-15T00:00:00Z", "2026-09-01T00:00:00Z"
	pv.data.projects.Projects = []tasksctl.Project{{Prefix: "zed", Root: "/r/zed", Reachable: true, LastActivity: &newer}, {Prefix: "abc", Root: "/r/abc", Reachable: true, LastActivity: &older}}
	pv.sortAndFilter()
	first := func(want, label string) {
		t.Helper()
		if pv.rows[0].Prefix != want {
			t.Fatalf("%s: %s first", label, pv.rows[0].Prefix)
		}
	}
	first("zed", "default activity desc")
	chord(t, d, app, "s", "p")
	first("abc", "s p")
	chord(t, d, app, "s", "p")
	first("abc", "repeated s p remains ascending")
	chord(t, d, app, "S", "p")
	first("zed", "S p")
	chord(t, d, app, "s", "a")
	first("abc", "s a")
	chord(t, d, app, "S", "a")
	first("zed", "S a")
}
