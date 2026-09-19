package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/khughitt/tasks-tui/internal/tasksctl"
	"github.com/khughitt/tasks-tui/internal/uitest"
)

func transitions(t *testing.T, f *fakeRunner) (*uitest.Driver, *App) {
	t.Helper()
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Expect("tui-aaa111")
	return d, app
}

func TestStartRunsInCheckoutAndReloads(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "start tui-aaa111", `{"id":"tui-aaa111","warnings":["took over stale claim sid:2"]}`)
	d, app := transitions(t, f)
	d.Key("space")
	if !logged(app, LevelWarning, "start tui-aaa111: took over stale claim sid:2") {
		t.Fatalf("start warning missing: %+v", app.Messages())
	}
	if !f.called("@/r/tui start tui-aaa111") {
		t.Fatal("start must run -C in the registered root when nothing is parked or claimed")
	}
	if n := countCalls(f, "ready --project tui"); n != 2 {
		t.Fatalf("the view must reload after a write; ready loaded %d times", n)
	}
}

func countCalls(f *fakeRunner, k string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c == k {
			n++
		}
	}
	return n
}

func TestStartClaimedOffersForce(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "start tui-aaa111", &tasksctl.Error{Kind: "claimed", Detail: "tui-aaa111 is claimed by sid:9 (feat/y)"})
	f.on("/r/tui", "start --force tui-aaa111", `{"id":"tui-aaa111","warnings":[]}`)
	d, app := transitions(t, f)
	d.Key("space")
	d.Expect("claimed by sid:9", "[F] start --force")
	d.Key("F")
	if !logged(app, LevelInfo, "start tui-aaa111") {
		t.Fatalf("start notice missing: %+v", app.Messages())
	}
	if !f.called("@/r/tui start --force tui-aaa111") {
		t.Fatal("F must run start --force")
	}
}

func TestStartOtherErrorIsShownNotForced(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "start tui-aaa111", &tasksctl.Error{Kind: "rename_pending", Detail: "project frozen"})
	d, _ := transitions(t, f)
	d.Key("space")
	d.Expect("start tui-aaa111: rename_pending: project frozen")
	d.ExpectNot("[F]")
}

func TestParkPromptTogglesAndRunsInParkedWorktree(t *testing.T) {
	f := projectFake()
	f.on("/wt/c", "park tui-ccc333 write the tests --waiting-on user --reason review", `{"id":"tui-ccc333","warnings":[]}`)
	d, app := transitions(t, f)
	d.Key("2")
	d.Expect("tui-ccc333")
	d.Key("j")
	d.Key("p")
	d.Expect("park tui-ccc333 — next step", "waiting on: agent", "reason: none")
	d.Type("write the tests")
	d.Key("ctrl+u")
	d.Expect("waiting on: user")
	d.Key("ctrl+r")
	d.Expect("reason: review")
	d.Key("enter")
	if !logged(app, LevelInfo, "park tui-ccc333") {
		t.Fatalf("park notice missing: %+v", app.Messages())
	}
	if !f.called("@/wt/c park tui-ccc333 write the tests --waiting-on user --reason review") {
		t.Fatalf("park argv/dir wrong: %v", f.writes())
	}
}

func TestPasteGoesToThePrompt(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "done tui-aaa111 pasted text", `{"id":"tui-aaa111","warnings":[]}`)
	d, _ := transitions(t, f)
	d.Key("d")
	d.Feed(tea.PasteMsg{Content: "pasted text"})
	d.Key("enter")
	if !f.called("@/r/tui done tui-aaa111 pasted text") {
		t.Fatalf("pasted text must reach the input: %v", f.writes())
	}
}

func TestParkRefusesEmptyNextStepAndEscCancels(t *testing.T) {
	f := projectFake()
	d, _ := transitions(t, f)
	d.Key("p")
	d.Expect("next step")
	d.Key("enter")
	d.Expect("next step is required")
	d.Key("esc")
	d.ExpectNot("next step is required")
	if w := f.writes(); len(w) != 0 {
		t.Fatalf("no write may run: %v", w)
	}
}

func TestDoneAndDropWithConfirm(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "done tui-aaa111 landed", `{"id":"tui-aaa111","warnings":[]}`)
	f.on("/r/tui", "drop tui-aaa111", `{"id":"tui-aaa111","warnings":[]}`)
	d, app := transitions(t, f)
	d.Key("d")
	d.Expect("done tui-aaa111 — what landed")
	d.Type("landed")
	d.Key("enter")
	if !logged(app, LevelInfo, "done tui-aaa111") {
		t.Fatalf("done notice missing: %+v", app.Messages())
	}
	if !f.called("@/r/tui done tui-aaa111 landed") {
		t.Fatalf("done argv: %v", f.writes())
	}
	d.Key("x")
	d.Expect("drop tui-aaa111 — why")
	d.Key("enter")
	d.Expect("really drop tui-aaa111", "[y] drop")
	d.Key("q")
	if app.overlay != nil || d.Quit || f.called("@/r/tui drop tui-aaa111") {
		t.Fatal("q cancels the confirm without writing or quitting")
	}
	d.Key("x")
	d.Key("enter")
	d.Expect("[y] drop")
	d.Key("y")
	if !logged(app, LevelInfo, "drop tui-aaa111") {
		t.Fatalf("drop notice missing: %+v", app.Messages())
	}
	if !f.called("@/r/tui drop tui-aaa111") {
		t.Fatalf("drop argv: %v", f.writes())
	}
}

func TestDoneRefusalIsShown(t *testing.T) {
	f := projectFake()
	f.on("/r/tui", "done tui-aaa111", &tasksctl.Error{Kind: "open_descendants", Detail: "2 children still open"})
	d, _ := transitions(t, f)
	d.Key("d")
	d.Key("enter")
	d.Expect("done tui-aaa111: open_descendants: 2 children still open")
}

func TestTransitionsNeedATarget(t *testing.T) {
	f := newFake()
	f.on("", "projects", projectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project tui", primeJSON("tui"))
	env := testEnv(f)
	d := drive(t, New(env, Options{Stack: []view{newProjectsView(env)}}))
	for _, k := range []string{"space", "p", "d", "x"} {
		d.Key(k)
		d.Expect("no task highlighted")
	}
	if w := f.writes(); len(w) != 0 {
		t.Fatalf("no write may run: %v", w)
	}
}
