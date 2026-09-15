package ui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/tasksctl"
)

func projectFake() *fakeRunner {
	f := newFake()
	f.on("/r/tui", "show tui-ddd444", showJSON(tasksctl.Task{ID: "tui-ddd444", Title: "Title of tui-ddd444", Status: "todo", Priority: 3, Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Depends: []string{}, Tags: []string{}, Body: "unique task body", Notes: []tasksctl.Note{}}, nil))
	f.on("", "prime --project tui", primeJSON("tui"))
	f.on("", "ready --project tui", rowsJSON(row("tui-aaa111", "todo", 2), row("tui-ddd444", "todo", 3)))
	f.on("", "list --project tui --status doing", rowsJSON(row("tui-bbb222", "doing", 1)))
	f.on("", "list --project tui --parked", parkedJSON(parkedRow("tui-ccc333", "parked one", "todo", "/wt/c")))
	f.on("", "list --project tui", rowsJSON(row("tui-aaa111", "todo", 2), row("tui-bbb222", "doing", 1)))
	f.on("", "list --project tui --status idea", rowsJSON(row("tui-eee555", "idea", 2)))
	f.on("", "list --project tui --status done --sort updated", rowsJSON(row("tui-fff666", "done", 2)))
	return f
}

func TestProjectViewTabsFilterAndCurrent(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	d.Expect("tui-aaa111", "tui-ddd444", "tui", "/r/tui", "1 doing", "2 todo", "1 idea", "0 blocked")
	if cur := pv.current(); cur == nil || cur.ID != "tui-aaa111" {
		t.Fatalf("current %+v", cur)
	}
	d.Key("j")
	if cur := pv.current(); cur == nil || cur.ID != "tui-ddd444" {
		t.Fatalf("after j current %+v", cur)
	}
	d.Key("tab")
	d.Expect("tui-bbb222", "tui-ccc333", "⏸", "user")
	d.Key("j")
	if cur := pv.current(); cur == nil || cur.ID != "tui-ccc333" || cur.Park == nil {
		t.Fatalf("parked row target %+v", cur)
	}
	d.Key("4")
	d.Expect("tui-eee555")
	d.Key("5")
	d.Expect("tui-fff666")
	d.Key("shift+tab")
	d.Expect("tui-eee555")
	d.Key("1")
	d.Expect("tui-ddd444")
	d.Key("/")
	d.Type("ddd")
	d.Key("enter")
	d.ExpectNot("tui-aaa111")
	if cur := pv.current(); cur == nil || cur.ID != "tui-ddd444" || len(pv.rows) != 1 {
		t.Fatalf("filter: rows=%d current=%+v", len(pv.rows), cur)
	}
	d.Key("enter")
	d.Expect("unique task body")
	if !f.called("ready --project tui") || !f.called("prime --project tui") {
		t.Fatal("ready and prime must have loaded")
	}
}

func TestProjectViewUnresolvedHasNoTarget(t *testing.T) {
	f := projectFake()
	f.on("", "list --project tui --parked", parkedJSON(tasksctl.ParkedRow{ID: "tui-ggg777", Title: "gone", Tags: []string{}, Park: &tasksctl.ParkInfo{At: "t", WaitingOn: "agent", NextStep: "n", Session: "s", Owner: "o", Host: "h", Worktree: "/wt/g"}}))
	f.on("", "list --project tui --status doing", rowsJSON())
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	d.Key("2")
	d.Expect("unresolved", "⏸", "agent")
	if pv.current() != nil {
		t.Fatal("an unresolved parked row offers no transitions")
	}
}

func TestProjectViewMergesParkedDoingRows(t *testing.T) {
	f := projectFake()
	f.on("", "list --project tui --parked", parkedJSON(parkedRow("tui-bbb222", "Title of tui-bbb222", "doing", "/wt/b"), parkedRow("tui-ccc333", "parked one", "todo", "/wt/c")))
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	d.Key("2")
	if len(pv.rows) != 2 || pv.current() == nil || pv.current().Park == nil || strings.Count(d.Screen(), "Title of tui-bbb222") != 1 {
		t.Fatalf("parked doing merge rows=%d current=%+v", len(pv.rows), pv.current())
	}
}

func TestProjectViewFilterUTF8AndSelectionPosition(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	d.Key("/")
	d.Type("sé")
	d.Key("backspace")
	if pv.filter != "s" {
		t.Fatalf("UTF-8 backspace left %q", pv.filter)
	}
	d.Key("esc")
	d.Key("esc")
	if d.Quit || len(f.writes()) != 0 {
		t.Fatalf("filter keys escaped to app: quit=%v writes=%v", d.Quit, f.writes())
	}
	pv.filter = ""
	pv.all = []rowView{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	pv.rows = append([]rowView(nil), pv.all...)
	pv.sel = 2
	pv.all = pv.all[:1]
	pv.applyFilter()
	if pv.sel != 0 {
		t.Fatalf("selection = %d, want clamped previous position", pv.sel)
	}
}

func TestProjectViewWarningSourcesAndBackgroundDrain(t *testing.T) {
	f := projectFake()
	f.on("", "prime --project tui", strings.Replace(primeJSON("tui"), "tui has a stale claim file", "prime warning", 1))
	f.on("", "list --project tui --status doing", `{"tasks":[],"warnings":["doing warning"]}`)
	f.on("", "list --project tui --parked", `{"tasks":[],"warnings":["parked warning"]}`)
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	d.Key("2")
	for _, want := range []string{"prime --project tui: prime warning", "list --project tui --status doing: doing warning", "list --project tui --parked: parked warning"} {
		if !logged(app, LevelWarning, want) {
			t.Fatalf("missing %q in %+v", want, app.Messages())
		}
	}
	pending := func(gen uint64) tea.Cmd { return func() tea.Msg { return gen } }
	before := len(pv.rows)
	pv.loader.Request(pending)
	gen := pv.loader.gen
	pv.loader.Request(pending)
	_, cmd := pv.update(loadMsg{loader: pv.loader.ID(), gen: gen, data: projectData{tab: pv.tab, rows: []rowView{{ID: "leak"}}}, err: errors.New("leak"), background: true})
	if cmd == nil || cmd() != uint64(gen+1) || len(pv.rows) != before || (len(pv.rows) > 0 && pv.rows[0].ID == "leak") {
		t.Fatal("background result must drain pending work without storing data or notices")
	}
}

func TestProjectViewHiddenResultIsDroppedThenPopReloads(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	pv.rows = []rowView{{ID: "kept"}}
	app := New(env, Options{Stack: []view{pv}})
	if pv.reload() == nil {
		t.Fatal("initial parent load must start")
	}
	gen := pv.loader.gen
	app.push(&stubView{name: "child"})
	app.Update(loadMsg{loader: pv.loader.ID(), gen: gen, data: projectData{tab: tabReady, rows: []rowView{{ID: "leak"}}}, err: errors.New("hidden failure")})
	if pv.rows[0].ID != "kept" || len(app.Messages()) != 0 {
		t.Fatalf("hidden parent result leaked: rows=%+v messages=%+v", pv.rows, app.Messages())
	}
	cmd := app.pop()
	if cmd == nil {
		t.Fatal("returning to the parent must start a reload after its hidden completion")
	}
	msg := cmd()
	if _, ok := msg.(loadMsg); !ok {
		t.Fatalf("reload returned %T, want loadMsg", msg)
	}
}

func TestProjectViewPillTabCountHeaderRowAndEmptyStates(t *testing.T) {
	f := projectFake()
	f.on("", "list --project tui --status idea", rowsJSON())
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	d.Expect(" 1 Ready 2 ", "status", "title", "  2 Doing")
	if pill := env.Styles.Pill(2).Render(" 1 Ready 2 "); !strings.Contains(app.render(), pill) {
		t.Fatalf("active tab is a pill: %q", app.render())
	}
	d.ExpectNot("2 rows")
	d.Key("/")
	d.Type("ddd")
	d.Expect("/ddd · 1 of 2", " 1 Ready 1 ")
	d.Key("esc")
	d.Key("/")
	for range 3 {
		d.Key("backspace")
	}
	d.Key("esc")
	d.Expect(" 1 Ready 2 ")
	d.Key("4")
	d.Expect("no ideas — a then ? files one", " 4 Ideas 0 ")
	d.Key("/")
	d.Type("zzz")
	d.Expect("no rows match /zzz")
	d.Key("esc")
	d.Key("/")
	for range 3 {
		d.Key("backspace")
	}
	d.Key("esc")
	d.Key("2")
	d.Expect(" 2 Doing 2 ")
	cmd := pv.setTab(tabReady)
	if !strings.Contains(app.render(), " 1 Ready 2 ") || !strings.Contains(app.render(), " 2 Doing ") {
		t.Fatalf("pending tab keeps its last count:\n%s", app.render())
	}
	d.Run(cmd)
	d.Expect(" 1 Ready 2 ")
}
