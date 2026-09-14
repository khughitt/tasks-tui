package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/tasksctl"
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
	d.Expect("tui", "ops ✗ unreachable", "1 doing · 1 parked across 2 projects", "tui-aaa111", "tui-bbb222")
	if !logged(app, LevelWarning, "projects: registry: ops unreachable") || !logged(app, LevelWarning, "prime --project tui: tui has a stale claim file") {
		t.Fatalf("warnings missing: %+v", app.Messages())
	}
	d.Key("S")
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
	if logged(app, LevelWarning, "prime --project ops") || pv.primeFor != "tui" {
		t.Fatalf("stale pane accepted: %+v", app.Messages())
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
	d.Expect("p6 ✗ unreachable")
	d.ExpectNot("p0 ✗ unreachable")
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

func TestProjectsBackgroundResultsAreDrainedAndDropped(t *testing.T) {
	env := testEnv(newFake())
	pv := newProjectsView(env)
	gen := pv.main.gen
	_, cmd := pv.update(loadMsg{loader: pv.main.ID(), gen: gen, data: projectsData{}, background: true})
	if cmd != nil || len(pv.data.projects.Projects) != 0 {
		t.Fatal("background main result accepted")
	}
	gen = pv.pane.gen
	_, cmd = pv.update(loadMsg{loader: pv.pane.ID(), gen: gen, data: paneData{prefix: "tui"}, background: true})
	if cmd != nil || pv.prime != nil {
		t.Fatal("background pane result accepted")
	}
}
