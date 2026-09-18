package ui

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/tasksctl"
)

func showJSON(t tasksctl.Task, extra map[string]any) string {
	m := map[string]any{"task": t, "spec_path": nil, "plan_path": nil, "step_found": nil, "depends_on": []any{},
		"parent": nil, "children": []any{}, "claim": nil, "park": nil, "escalation": nil, "periodic": nil, "warnings": []string{}}
	for k, v := range extra {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func TestTaskViewShowsFromCheckoutAndRendersBody(t *testing.T) {
	f := newFake()
	size := "m"
	spec := "docs/specs/x.md"
	task := tasksctl.Task{ID: "tui-aaa111", Title: "Do the thing", Status: "todo", Priority: 2, Size: &size,
		Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Depends: []string{}, Tags: []string{"ui"}, Spec: &spec,
		Body: "# Why\n\nBecause **bold** reasons.", Notes: []tasksctl.Note{{At: "2026-09-13T10:00:00Z", By: "main", Text: "a note"}}}
	specPath := "/wt/a/docs/specs/x.md"
	f.on("/wt/a", "show tui-aaa111", showJSON(task, map[string]any{
		"spec_path":  specPath,
		"park":       tasksctl.ParkInfo{At: "2026-09-13T10:00:00Z", NextStep: "resume here", WaitingOn: "user", Session: "sid:1", Owner: "feat/a", Host: "h", Worktree: "/wt/a"},
		"claim":      tasksctl.ClaimInfo{Owner: "feat/a", Session: "sid:1", Host: "h", Worktree: "/wt/a", Started: "t", Seen: "t", Live: false},
		"depends_on": []map[string]any{{"id": "tui-000000", "title": "dep", "status": "done", "resolved": true}, {"id": "ops-999999", "title": nil, "status": nil, "resolved": false}},
		"warnings":   []string{"worktree copies diverge"},
	}))
	env := testEnv(f)
	tv := newTaskView(env, target{ID: "tui-aaa111", Title: "Do the thing", Prefix: "tui", Park: &tasksctl.ParkInfo{Worktree: "/wt/a"}})
	app := New(env, Options{Stack: []view{tv}})
	d := drive(t, app)
	d.Expect("Do the thing", "P2", "Because", "bold", "a note", "resume here", "waiting on user", "stale", "feat/a",
		"tui-000000", "ops-999999", "not reachable from here", "docs/specs/x.md")
	if got := tv.warnings(); !slices.Equal(got, []string{"show tui-aaa111: worktree copies diverge"}) || len(app.Messages()) != 0 {
		t.Fatalf("show warnings are the view's, not the log's: %q %+v", got, app.Messages())
	}
	d.Expect("⚠ 1")
	if !f.called("@/wt/a show tui-aaa111") {
		t.Fatal("show must run in the parked checkout")
	}
	if cur := tv.current(); cur == nil || cur.ID != "tui-aaa111" || cur.Park == nil {
		t.Fatalf("current %+v", cur)
	}
	d.Key("j")
	d.Key("pgdown")
	chord(t, d, app, "g", "g")
	d.Expect("Do the thing")
}

func TestTaskViewReportsMissingWorktreeNotice(t *testing.T) {
	f := newFake()
	task := tasksctl.Task{ID: "tui-aaa111", Title: "T", Status: "todo", Priority: 3, Created: "2026-09-13T09:00:00Z", Updated: "2026-09-13T10:00:00Z", Depends: []string{}, Tags: []string{}, Notes: []tasksctl.Note{}}
	f.on("/r/tui", "show tui-aaa111", showJSON(task, nil))
	env := testEnv(f)
	tv := newTaskView(env, target{ID: "tui-aaa111", Prefix: "tui", Park: &tasksctl.ParkInfo{Worktree: "/gone"}})
	app := New(env, Options{Stack: []view{tv}})
	d := drive(t, app)
	d.Expect("⚠ 1")
	d.Key("W")
	d.Expect("parked checkout /gone is gone; using /r/tui")
	if got := tv.warnings(); len(got) != 1 || !strings.Contains(got[0], "parked checkout /gone is gone") {
		t.Fatalf("the checkout notice is a load warning: %q", got)
	}
	if !f.called("@/r/tui show tui-aaa111") {
		t.Fatal("show must fall back to the registered root")
	}
}

func TestTaskViewDropsBackgroundResultAndDrainsPendingWork(t *testing.T) {
	tv := newTaskView(testEnv(newFake()), target{ID: "tui-aaa111", Prefix: "tui"})
	pending := func(gen uint64) tea.Cmd { return func() tea.Msg { return gen } }
	tv.loader.Request(pending)
	gen := tv.loader.gen
	tv.loader.Request(pending)
	_, cmd := tv.update(loadMsg{loader: tv.loader.ID(), gen: gen, data: taskData{}, err: errors.New("hidden failure"), background: true})
	if cmd == nil || cmd() != uint64(gen+1) || tv.data != nil {
		t.Fatal("background result must drain pending work without storing data or notices")
	}
}

func TestTaskViewUsesSlotAccentForDoingAndLiveClaim(t *testing.T) {
	env := testEnv(newFake())
	tv := newTaskView(env, target{ID: "tui-aaa111", Prefix: "tui"})
	tv.data = &taskData{res: tasksctl.ShowResult{
		Task:  tasksctl.Task{ID: "tui-aaa111", Title: "T", Status: "doing", Priority: 1, Depends: []string{}, Tags: []string{}, Notes: []tasksctl.Note{}},
		Claim: &tasksctl.ClaimInfo{Owner: "feat/a", Live: true},
	}}
	wantDoing := strings.TrimSuffix(env.Styles.Accent(2).Bold(true).Render("doing"), "\x1b[m")
	wantClaim := strings.TrimSuffix(env.Styles.Accent(2).Bold(true).Render("live"), "\x1b[m")
	content := tv.content()
	if !strings.Contains(content, wantDoing) || !strings.Contains(content, wantClaim) {
		t.Fatalf("doing and live claim must use slot accent:\n%s", content)
	}
}
