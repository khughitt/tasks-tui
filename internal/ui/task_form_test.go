package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"tasks-tui/internal/tasksctl"
)

func TestEditFormPrefillsAndSavesThreeFields(t *testing.T) {
	f := projectFake()
	task := tasksctl.Task{ID: "tui-aaa111", Title: "Old title", Status: "todo", Priority: 2,
		Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Depends: []string{},
		Tags: []string{"old"}, Body: "Old body", Notes: []tasksctl.Note{}}
	f.on("/r/tui", "show tui-aaa111", showJSON(task, nil))
	f.on("/r/tui", "edit tui-aaa111 --title New title -b New body --no-tags --tag new --tag second", `{"id":"tui-aaa111","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)

	d.Key("e")
	d.Expect("edit tui-aaa111", "title", "Old title", "body", "Old body", "tags", "old")
	form, ok := app.top().(*taskFormView)
	if !ok || !form.titleInput.Focused() {
		t.Fatalf("edit must open with title focused: %T", app.top())
	}
	screen := d.Screen()
	if !(strings.Index(screen, "title") < strings.Index(screen, "body") && strings.Index(screen, "body") < strings.Index(screen, "tags")) {
		t.Fatalf("fields are out of order:\n%s", screen)
	}
	frame := strings.Index(screen, "╭")
	if frame < 1 || !strings.Contains(screen[:frame], "\n") {
		t.Fatalf("form is not vertically centered:\n%s", screen)
	}

	d.Key("ctrl+u")
	d.Type("New title")
	d.Key("tab")
	d.Key("ctrl+u")
	d.Type("New body")
	d.Key("tab")
	d.Key("ctrl+u")
	d.Type("new second")
	d.Key("ctrl+s")

	if !f.called("@/r/tui edit tui-aaa111 --title New title -b New body --no-tags --tag new --tag second") {
		t.Fatalf("edit argv: %v", f.calls)
	}
	if _, open := app.top().(*taskFormView); open {
		t.Fatal("successful save must close the form")
	}
}

func TestAddFormUsesTheSameThreeFields(t *testing.T) {
	f := projectFake()
	f.on("", "add New task --project tui -b First line\nSecond line --tag ui --tag docs", `{"id":"tui-new001","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)

	d.Key("a")
	d.Expect("add task", "title", "body", "tags")
	form, ok := app.top().(*taskFormView)
	if !ok || !form.titleInput.Focused() {
		t.Fatalf("add must use the task form with title focused: %T", app.top())
	}
	d.Type("New task")
	d.Key("tab")
	d.Type("First line")
	d.Key("enter")
	d.Type("Second line")
	d.Key("tab")
	d.Type("ui docs")
	d.Key("ctrl+s")

	if !f.called("add New task --project tui -b First line\nSecond line --tag ui --tag docs") {
		t.Fatalf("add argv: %v", f.calls)
	}
	if !logged(app, LevelInfo, "created tui-new001") {
		t.Fatalf("created task was not logged: %v", app.Messages())
	}
}

func TestTaskFormCompletesProjectTags(t *testing.T) {
	f := projectFake()
	f.on("", "tags --project tui", `{"tags":[{"tag":"backend","meaning":null,"count":2,"projects":{"tui":2}},{"tag":"bug","meaning":null,"count":1,"projects":{"tui":1}}],"warnings":[]}`)
	f.on("", "add Note --project tui --tag bug", `{"id":"tui-new002","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)

	d.Key("a")
	d.Type("Note")
	d.Key("tab")
	d.Key("tab")
	d.ExpectNot("backend")
	d.Type("b")
	d.Expect("backend")
	d.Key("down")
	d.Key("enter")
	d.Key("ctrl+s")

	if !f.called("add Note --project tui --tag bug") {
		t.Fatalf("completed tag was not submitted: %v", f.calls)
	}
}

func TestTaskFormRequiresTitleAndKeepsSubmissionErrorsOpen(t *testing.T) {
	f := projectFake()
	f.on("", "add Note --project tui", &tasksctl.Error{Kind: "write", Detail: "offline"})
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)

	d.Key("a")
	d.Key("ctrl+s")
	d.Expect("title is required")
	if f.called("add  --project tui") {
		t.Fatal("blank title must not be submitted")
	}
	d.Type("Note")
	d.Key("ctrl+s")
	d.Expect("add task: write: offline")
	if _, open := app.top().(*taskFormView); !open {
		t.Fatal("a failed submission must keep the form open")
	}
}

func TestEditFormDoesNotSubmitBeforeTaskLoads(t *testing.T) {
	form := newTaskFormView(testEnv(newFake()), target{ID: "tui-aaa111", Prefix: "tui"})
	form.titleInput.SetValue("Changed")
	if cmd := form.submit(); cmd != nil || form.problem != "task is still loading" {
		t.Fatalf("submit while loading: cmd=%v problem=%q", cmd != nil, form.problem)
	}
}

func TestEditFormShowsLoadErrors(t *testing.T) {
	f := newFake()
	f.on("/r/tui", "show tui-aaa111", &tasksctl.Error{Kind: "lookup", Detail: "offline"})
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newTaskFormView(env, target{ID: "tui-aaa111", Prefix: "tui"})}})
	d := drive(t, app)
	d.Expect("edit tui-aaa111: lookup: offline")
}

func TestClosingTaskFormReloadsTheUnderlyingView(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("a")
	d.Key("esc")
	if got := countCalls(f, "ready --project tui"); got != 2 {
		t.Fatalf("ready reloads after close: got %d, want 2", got)
	}
}

func TestRefreshDoesNotOverwriteAnOpenEditForm(t *testing.T) {
	f := projectFake()
	task := tasksctl.Task{ID: "tui-aaa111", Title: "Old", Status: "todo", Priority: 2,
		Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Body: "", Tags: []string{}}
	f.on("/r/tui", "show tui-aaa111", showJSON(task, nil))
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("e")
	d.Key("ctrl+u")
	d.Type("Unsaved")
	d.Feed(tickMsg{})
	form := app.top().(*taskFormView)
	if got := form.titleInput.Value(); got != "Unsaved" {
		t.Fatalf("refresh overwrote title: %q", got)
	}
	if got := countCalls(f, "@/r/tui show tui-aaa111"); got != 1 {
		t.Fatalf("show calls while form open: got %d, want 1", got)
	}
}

func TestTaskFormAllowsOnlyOneSubmissionAtATime(t *testing.T) {
	f := newFake()
	f.on("", "add Note --project tui", &tasksctl.Error{Kind: "write", Detail: "offline"})
	form := newAddFormView(testEnv(f), "tui")
	form.titleInput.SetValue("Note")
	_, first := form.update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if first == nil {
		t.Fatal("first save did not start")
	}
	if _, second := form.update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}); second != nil {
		t.Fatal("second save started while the first was in flight")
	}
	if _, cancel := form.update(tea.KeyPressMsg{Code: tea.KeyEscape}); cancel != nil {
		t.Fatal("escape closed the form while its write was in flight")
	}
	form.update(first())
	if _, retry := form.update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}); retry == nil {
		t.Fatal("failed save did not allow a retry")
	}
}

func TestEditFormReportsCheckoutAndShowWarnings(t *testing.T) {
	f := newFake()
	task := tasksctl.Task{ID: "tui-aaa111", Title: "T", Status: "todo", Priority: 2,
		Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Body: "", Tags: []string{}}
	f.on("/r/tui", "show tui-aaa111", showJSON(task, map[string]any{"warnings": []string{"copies diverge"}}))
	f.on("", "tags --project tui", `{"tags":[],"warnings":[]}`)
	env := testEnv(f)
	form := newTaskFormView(env, target{ID: "tui-aaa111", Prefix: "tui", Park: &tasksctl.ParkInfo{Worktree: "/gone"}})
	app := New(env, Options{Stack: []view{form}})
	drive(t, app)
	if !logged(app, LevelWarning, "parked checkout /gone is gone; using /r/tui") || !logged(app, LevelWarning, "show tui-aaa111: copies diverge") {
		t.Fatalf("edit warnings missing: %v", app.Messages())
	}
}
