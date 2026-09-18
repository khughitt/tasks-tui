package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"tasks-tui/internal/tasksctl"
)

func TestQuickAddCompletesTagAndSubmitsIt(t *testing.T) {
	f := projectFake()
	f.on("", "tags --project ops", `{"tags":[{"tag":"backend","meaning":null,"count":2,"projects":{"ops":2}},{"tag":"bug","meaning":null,"count":1,"projects":{"ops":1}}],"warnings":["tag notice"]}`)
	f.on("", "add 修正 --project ops --tag bug", `{"id":"ops-new001","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("a")
	d.Feed(tea.PasteMsg{Content: "修正 >ops #b"})
	d.Expect("backend")
	d.Key("down")
	d.Key("tab")
	d.Expect("--tag bug")
	d.Key("enter")
	if !f.called("add 修正 --project ops --tag bug") || !logged(app, LevelInfo, "created ops-new001") {
		t.Fatalf("completed tag wasn't submitted: calls=%v messages=%v", f.calls, app.Messages())
	}
}

// Complete commands but withhold their load message, modeling a delayed delivery.
func heldTagResult(t *testing.T, cmd tea.Cmd) loadMsg {
	t.Helper()
	var found []loadMsg
	var run func(tea.Cmd)
	run = func(cmd tea.Cmd) {
		if cmd == nil {
			return
		}
		switch msg := cmd().(type) {
		case tea.BatchMsg:
			for _, child := range msg {
				run(child)
			}
		case loadMsg:
			found = append(found, msg)
		}
	}
	run(cmd)
	if len(found) != 1 {
		t.Fatalf("wanted one pending tag result, got %d", len(found))
	}
	return found[0]
}

func TestQuickAddDropsOldTagResults(t *testing.T) {
	for _, reopen := range []bool{false, true} {
		for _, failed := range []bool{false, true} {
			t.Run(fmt.Sprintf("reopen=%t/error=%t", reopen, failed), func(t *testing.T) {
				f := projectFake()
				f.on("", "tags --project tui", `{"tags":[{"tag":"backend","meaning":null,"count":1,"projects":{"tui":1}}],"warnings":["old warning"]}`)
				if failed {
					f.on("", "tags --project tui", &tasksctl.Error{Kind: "lookup", Detail: "old failure"})
				}
				f.on("", "tags --project ops", `{"tags":[{"tag":"operations","meaning":null,"count":1,"projects":{"ops":1}}],"warnings":[]}`)
				env := testEnv(f)
				app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
				d := drive(t, app)
				d.Key("a")
				_, pending := app.Update(tea.PasteMsg{Content: "note #b"})
				old := heldTagResult(t, pending)
				if reopen {
					d.Key("esc")
					d.Key("a")
				} else {
					d.Key("ctrl+u")
				}
				d.Feed(tea.PasteMsg{Content: "note >ops #o"})
				d.Feed(old)
				d.Expect("operations")
				d.ExpectNot("backend")
				if logged(app, LevelError, "old failure") || logged(app, LevelWarning, "old warning") {
					t.Fatalf("stale lookup reached log: %v", app.Messages())
				}
				d.Key("tab")
				d.Expect("--tag operations")
			})
		}
	}
}

func TestQuickAddAllowsNewTagsWithoutSuggestions(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(fmt.Sprintf("lookupFailure=%t", failed), func(t *testing.T) {
			f := projectFake()
			f.on("", "tags --project tui", `{"tags":[],"warnings":[]}`)
			if failed {
				f.on("", "tags --project tui", &tasksctl.Error{Kind: "lookup", Detail: "offline"})
			}
			f.on("", "add note --project tui --tag new", `{"id":"tui-new001","action":"created","warnings":[]}`)
			env := testEnv(f)
			app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
			d := drive(t, app)
			d.Key("a")
			d.Feed(tea.PasteMsg{Content: "note #new"})
			if failed && !logged(app, LevelError, "tags --project tui: lookup: offline") {
				t.Fatalf("lookup error missing: %v", app.Messages())
			}
			d.Key("tab")
			d.Key("enter")
			if !f.called("add note --project tui --tag new") {
				t.Fatalf("manual tag blocked: %v", f.calls)
			}
		})
	}
}

func TestQuickAddDropsOldTagsAfterReturningToSameProject(t *testing.T) {
	f := projectFake()
	f.on("", "tags --project tui", `{"tags":[{"tag":"backend","meaning":null,"count":1,"projects":{"tui":1}}],"warnings":["old snapshot"]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("a")
	_, cmd := app.Update(tea.PasteMsg{Content: "note #b"})
	old := heldTagResult(t, cmd)
	d.Key("ctrl+u")
	d.Feed(tea.PasteMsg{Content: "note >ops #o"})
	d.Key("ctrl+u")
	d.Feed(tea.PasteMsg{Content: "note #b"})
	f.on("", "tags --project tui", `{"tags":[{"tag":"build","meaning":null,"count":1,"projects":{"tui":1}}],"warnings":[]}`)
	d.Feed(old)
	d.Expect("build")
	d.ExpectNot("backend")
	if logged(app, LevelWarning, "old snapshot") || f.called("tags --project ops") {
		t.Fatalf("superseded lookup was accepted or not coalesced: messages=%v calls=%v", app.Messages(), f.calls)
	}
}
