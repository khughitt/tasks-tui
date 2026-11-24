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
	// Non-request user park (environment) keeps the exact pre-Stage-B argv.
	f.on("/wt/c", "park tui-ccc333 fix the env --waiting-on user --reason environment", `{"id":"tui-ccc333","warnings":[]}`)
	d, app := transitions(t, f)
	d.Key("2")
	d.Expect("tui-ccc333")
	d.Key("j")
	d.Key("p")
	d.Expect("park tui-ccc333 — next step", "waiting on: agent", "reason: none")
	d.Type("fix the env")
	d.Key("ctrl+u")
	d.Expect("waiting on: user")
	for range 4 {
		d.Key("ctrl+r")
	}
	d.Expect("reason: environment")
	d.Key("enter")
	if !logged(app, LevelInfo, "park tui-ccc333") {
		t.Fatalf("park notice missing: %+v", app.Messages())
	}
	if !f.called("@/wt/c park tui-ccc333 fix the env --waiting-on user --reason environment") {
		t.Fatalf("park argv/dir wrong: %v", f.writes())
	}
}

func TestParkHumanRequestLinksExistingOpenDecision(t *testing.T) {
	const d1 = "6f76bdab-e417-4ae7-81ce-09fc70d6487a"
	f := projectFake()
	f.on("/wt/c", "show tui-ccc333", `{
		"task":{"id":"tui-ccc333","title":"C","status":"doing","priority":2,"parallel":false,"created":"2026-11-24T00:00:00Z","updated":"2026-11-24T00:00:00Z","body":""},
		"decisions":[{
			"id":"6f76bdab-e417-4ae7-81ce-09fc70d6487a","type":"decision","at":"2026-11-24T00:00:00Z","by":"keith",
			"subject":{"scope":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			"data":{"kind":"question","origin":"scope","text":"Review round 2?"},
			"state":"open","resolution":null,"seen_scope":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","scope_changed_since_open":false,
			"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T00:00:00Z"}
		}],
		"warnings":[]
	}`)
	f.on("/wt/c", "park tui-ccc333 open sheet --waiting-on user --reason review --decision "+d1, `{"id":"tui-ccc333","warnings":[]}`)
	d, app := transitions(t, f)
	d.Key("2")
	d.Expect("tui-ccc333")
	d.Key("j")
	d.Key("p")
	d.Type("open sheet")
	d.Key("ctrl+u")
	d.Key("ctrl+r")
	d.Expect("reason: review")
	d.Key("enter")
	d.Expect("link open decision", "question 6f76bdab: Review round 2?")
	d.Key("1")
	if !logged(app, LevelInfo, "park tui-ccc333") {
		t.Fatalf("park notice missing: %+v", app.Messages())
	}
	if !f.called("@/wt/c park tui-ccc333 open sheet --waiting-on user --reason review --decision " + d1) {
		t.Fatalf("linked park argv wrong: %v", f.writes())
	}

	// Selecting "new inline decision" from a non-empty open-decision picker opens the inline prompt.
	const dNew = "2c5f39cb-3fb2-4e3f-8d94-3e4f0d1b2c3d"
	orig := generateUUID
	generateUUID = func() string { return dNew }
	defer func() { generateUUID = orig }()
	wantNew := "park tui-ccc333 open sheet 2 --waiting-on user --reason review --decision " + dNew + " --decision-kind question --decision-origin tasks-tui --decision-text Fresh review?"
	f.on("/wt/c", wantNew, `{"id":"tui-ccc333","warnings":[]}`)
	d.Key("p")
	d.Type("open sheet 2")
	d.Key("ctrl+u")
	d.Key("ctrl+r")
	d.Expect("reason: review")
	d.Key("enter")
	d.Expect("link open decision", "new inline decision")
	d.Key("2")
	d.Expect("decision question")
	d.Type("Fresh review?")
	d.Key("enter")
	if !f.called(fakeKey("/wt/c", wantNew)) {
		t.Fatalf("inline decision from picker argv wrong: %v", f.writes())
	}
}

func TestParkHumanRequestInlineQuestionRetainsUUIDOnRetryAndApprovalRequiresPin(t *testing.T) {
	const fixedUUID = "1b4e28ba-2fa1-4d2f-9c83-2d3f9c0a1b2c"
	const sha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	orig := generateUUID
	generateUUID = func() string { return fixedUUID }
	defer func() { generateUUID = orig }()

	// 1. Inline question with failure + retry retains the exact generated UUID.
	f := projectFake()
	emptyShow := `{"task":{"id":"tui-aaa111","title":"A","status":"todo","priority":1,"parallel":false,"created":"2026-11-24T00:00:00Z","updated":"2026-11-24T00:00:00Z","body":""},"warnings":[]}`
	f.on("/r/tui", "show tui-aaa111", emptyShow)
	wantQ := "park tui-aaa111 wait on choice --waiting-on user --reason decision --decision " + fixedUUID + " --decision-kind question --decision-origin tasks-tui --decision-text Pick A or B?"
	f.on("/r/tui", wantQ, &tasksctl.Error{Kind: "io", Detail: "transient lock timeout"})
	d, app := transitions(t, f)
	d.Key("p")
	d.Type("wait on choice")
	d.Key("ctrl+u")
	d.Key("ctrl+r")
	d.Key("ctrl+r")
	d.Expect("reason: decision")
	d.Key("enter")
	d.Expect("decision question")
	d.Type("Pick A or B?")
	d.Key("enter")
	d.Expect("[r] retry park", "1b4e28ba")
	// Clear error on fake so retry succeeds with the EXACT same UUID!
	f.mu.Lock()
	delete(f.errors, fakeKey("/r/tui", wantQ))
	f.replies[fakeKey("/r/tui", wantQ)] = `{"id":"tui-aaa111","warnings":[]}`
	f.mu.Unlock()
	d.Key("r")
	if !logged(app, LevelInfo, "park tui-aaa111") {
		t.Fatalf("retry park did not succeed: %+v", app.Messages())
	}
	if n := countCalls(f, fakeKey("/r/tui", wantQ)); n != 2 {
		t.Fatalf("expected 2 calls with identical UUID on retry, got %d (%v)", n, f.writes())
	}

	// 2. Inline approval requires a valid pin (rejects incomplete pin, accepts digest pin).
	wantA := "park tui-aaa111 wait approval --waiting-on user --reason approval --decision " + fixedUUID + " --decision-kind approval --decision-origin tasks-tui --decision-text Approve spec? --artifact-path docs/specs/s.md --artifact-sha256 " + sha
	f.on("/r/tui", wantA, `{"id":"tui-aaa111","warnings":[]}`)
	d.Key("p")
	d.Type("wait approval")
	d.Key("ctrl+u")
	d.Key("ctrl+r")
	d.Key("ctrl+r")
	d.Key("ctrl+r")
	d.Expect("reason: approval")
	d.Key("enter")
	d.Expect("approval request")
	d.Type("Approve spec?")
	d.Key("enter")
	d.Expect("artifact pin")
	d.Type("docs/specs/s.md")
	d.Key("enter")
	d.Expect("use repo:path@commit")
	d.Type("@" + sha)
	d.Key("enter")
	if !f.called(fakeKey("/r/tui", wantA)) {
		t.Fatalf("approval park argv wrong: %v", f.writes())
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
