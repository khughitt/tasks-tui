package ui

import (
	"strings"
	"testing"

	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

const (
	testD1    = "6f76bdab-e417-4ae7-81ce-09fc70d6487a"
	testD2    = "1b4e28ba-2fa1-4d2f-9c83-2d3f9c0a1b2c"
	testD3    = "3c5e7a90-1234-4abc-8def-0123456789ab"
	testScope = "36ac255782a1ea3ca1796748e63071703ddd9b98dc740b8bd02ebaf8b3cfc444"
)

const goldenPendingFixture = `{
	"observed_at":"2026-11-24T17:07:06Z",
	"coverage":{"complete":false,"unavailable_projects":["beta"]},
	"items":[
		{
			"type":"decision",
			"task":{"id":"tui-aaa111","title":"Dual task","status":"todo","project":"tui"},
			"opened_at":"2026-11-24T16:00:00Z","age_seconds":4026,
			"decision":{
				"id":"6f76bdab-e417-4ae7-81ce-09fc70d6487a","type":"decision","at":"2026-11-24T16:00:00Z","by":"keith",
				"subject":{"scope":"9dffaa855b7cc25be8ee58ebe0a7e86fa076f63a7daa7c11e4fd91df190658b8"},
				"data":{"kind":"drop","origin":"scope","text":"Drop redundant task?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":"36ac255782a1ea3ca1796748e63071703ddd9b98dc740b8bd02ebaf8b3cfc444","scope_changed_since_open":true,
				"open_dependents":{"tasks":[{"id":"ops-123456","title":"Downstream"}],"complete":false,"unavailable_projects":["beta"],"observed_at":"2026-11-24T17:07:06Z"}
			}
		},
		{
			"type":"decision",
			"task":{"id":"tui-aaa111","title":"Dual task","status":"todo","project":"tui"},
			"opened_at":"2026-11-24T16:30:00Z","age_seconds":2226,
			"decision":{
				"id":"1b4e28ba-2fa1-4d2f-9c83-2d3f9c0a1b2c","type":"decision","at":"2026-11-24T16:30:00Z","by":"keith",
				"subject":{"repo":"records","path":"tasks/docs/plans/p.md","commit":"0123456789abcdef0123456789abcdef01234567"},
				"data":{"kind":"approval","origin":"workflow","text":"Approve plan rev 1?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":null,"scope_changed_since_open":null,
				"open_dependents":{"tasks":[],"complete":false,"unavailable_projects":["beta"],"observed_at":"2026-11-24T17:07:06Z"}
			}
		},
		{
			"type":"legacy_wait",
			"task":{"id":"tui-bbb222","title":"Legacy idea","status":"idea","project":"tui"},
			"opened_at":null,"age_seconds":null,
			"source":"awaiting_user","host":null,"reason":null,"text":"Keep legacy mode?"
		}
	],
	"warnings":[]
}`

const decAnswerOK = `{
	"id":"tui-aaa111","action":"answered","replayed":false,
	"decision":{
		"id":"6f76bdab-e417-4ae7-81ce-09fc70d6487a","type":"decision","at":"2026-11-24T16:00:00Z","by":"keith",
		"subject":{"scope":"9dffaa855b7cc25be8ee58ebe0a7e86fa076f63a7daa7c11e4fd91df190658b8"},
		"data":{"kind":"drop","origin":"scope","text":"Drop redundant task?","source":"direct"},
		"state":"resolved","resolution":null,"seen_scope":null,"scope_changed_since_open":null,
		"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T17:07:06Z"}
	},
	"resolution":null,"warnings":[]
}`

func TestPendingViewRendersSeparateRequestsScopeDepsCoverageAndLegacyNoAnswer(t *testing.T) {
	f := newFake()
	f.on("", "list --all-projects --pending", goldenPendingFixture)
	f.on("/r/tui", "show tui-bbb222", showJSON(tasksctl.Task{
		ID: "tui-bbb222", Title: "Legacy idea", Status: "idea", Priority: 2,
		Created: "2026-10-01T00:00:00Z", Updated: "2026-10-01T00:00:00Z", Body: "Legacy body.",
	}, nil))
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newPendingView(env, "")}})
	d := drive(t, app)

	// Separate rows for two decisions on tui-aaa111 + null age + changed scope + deps + incomplete coverage + legacy row.
	d.Expect(
		"incomplete coverage: beta",
		"6f76bdab", "scope CHANGED since open", "1 open dep(s): ops-123456 (Downstream)",
		"1b4e28ba", "pin records:tasks/docs/plans/p.md@01234567",
		"legacy:awaiting_user", "age unknown", "legacy evidence only; no direct answer",
	)
	if strings.Count(d.Screen(), "tui-aaa111") != 2 {
		t.Fatalf("expected two separate rows for two decisions on tui-aaa111:\n%s", d.Screen())
	}

	// Selecting the legacy row opens task detail (NO answer overlay and NO decision write).
	d.Key("j")
	d.Key("j")
	d.Key("enter")
	d.Expect("Legacy body.")
	if app.overlay != nil {
		t.Fatal("legacy wait row must not open a decision answer overlay")
	}
	if len(f.writes()) != 1 || f.writes()[0] != "@/r/tui show tui-bbb222" {
		t.Fatalf("expected only read-only show on legacy row, got %v", f.writes())
	}
}

func TestDecisionInteractiveDropAcceptCancelAndStaleScopeRefresh(t *testing.T) {
	f := newFake()
	f.on("", "list --project tui --pending", goldenPendingFixture)
	wantDropAccept := "decision answer tui-aaa111 " + testD1 + " --accept Drop it --seen-scope " + testScope
	f.on("/r/tui", wantDropAccept, decAnswerOK)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newPendingView(env, "tui")}})
	d := drive(t, app)
	d.Expect("6f76bdab")

	// 1. Open drop decision -> shows exact effect before sending; esc cancels without sending anything!
	d.Key("enter")
	d.Expect("accept (DROPS tui-aaa111 atomically; scope 36ac2557)", "decline (keeps task open)", "withdraw (cancels decision)")
	d.Key("1")
	d.Expect("effect: atomically drops tui-aaa111")
	d.Type("Drop it")
	d.Key("esc")
	if len(f.writes()) != 0 {
		t.Fatalf("cancellation must send no write, got %v", f.writes())
	}

	// 2. Submit drop accept -> passes exact displayed seen_scope token and reloads!
	d.Key("enter")
	d.Key("1")
	d.Type("Drop it")
	d.Key("enter")
	if !f.called(fakeKey("/r/tui", wantDropAccept)) {
		t.Fatalf("expected drop accept with displayed seen_scope, got %v", f.writes())
	}
	if !logged(app, LevelInfo, "decision answer tui-aaa111 (6f76bdab)") {
		t.Fatalf("missing answer notice: %+v", app.Messages())
	}

	// 3. Stale scope refusal logs warning and reloads WITHOUT automatic retry!
	f2 := newFake()
	f2.on("", "list --project tui --pending", goldenPendingFixture)
	f2.on("/r/tui", wantDropAccept, &tasksctl.Error{Kind: "stale_scope", Detail: "title edited"})
	env2 := testEnv(f2)
	app2 := New(env2, Options{Stack: []view{newPendingView(env2, "tui")}})
	d2 := drive(t, app2)
	d2.Expect("6f76bdab")
	d2.Key("enter")
	d2.Key("1")
	d2.Type("Drop it")
	d2.Key("enter")
	if !logged(app2, LevelWarning, "scope changed since displayed; refreshed") {
		t.Fatalf("expected stale_scope refresh warning, got %+v", app2.Messages())
	}
	if n := countCalls(f2, fakeKey("/r/tui", wantDropAccept)); n != 1 {
		t.Fatalf("stale_scope must never auto-retry, got %d calls", n)
	}
	if n := countCalls(f2, "list --project tui --pending"); n != 2 {
		t.Fatalf("stale_scope must reload pending view, got %d loads", n)
	}
}

func TestTaskViewDecisionSelectorHandlesQuestionReplyAndWithdraw(t *testing.T) {
	f := newFake()
	showWithTwo := `{
		"task":{"id":"tui-aaa111","title":"Dual","status":"todo","priority":2,"parallel":false,"created":"2026-11-24T00:00:00Z","updated":"2026-11-24T00:00:00Z","body":"Body"},
		"decisions":[
			{
				"id":"` + testD2 + `","type":"decision","at":"2026-11-24T16:00:00Z","by":"keith",
				"subject":{"scope":"` + testScope + `"},
				"data":{"kind":"question","origin":"scope","text":"Use v2 schema?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":"` + testScope + `","scope_changed_since_open":false,
				"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T17:00:00Z"}
			},
			{
				"id":"` + testD3 + `","type":"decision","at":"2026-11-24T16:05:00Z","by":"keith",
				"subject":{"scope":"` + testScope + `"},
				"data":{"kind":"suggestion","origin":"curate","text":"Close as delivered?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":"` + testScope + `","scope_changed_since_open":false,
				"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T17:00:00Z"}
			}
		],
		"warnings":[]
	}`
	f.on("/r/tui", "show tui-aaa111", showWithTwo)
	f.on("/r/tui", "decision answer tui-aaa111 "+testD2+" --reply Yes use v2", decAnswerOK)
	f.on("/r/tui", "decision withdraw tui-aaa111 "+testD3+" --reason Superseded", decAnswerOK)
	env := testEnv(f)
	tv := newTaskView(env, target{ID: "tui-aaa111", Prefix: "tui"})
	app := New(env, Options{Stack: []view{tv}})
	d := drive(t, app)
	d.Expect("decisions", "Use v2 schema?", "Close as delivered?", "enter: answer or withdraw")

	// Pick question #1 -> reply.
	d.Key("enter")
	d.Expect("select decision on tui-aaa111", "question 1b4e28ba", "suggestion 3c5e7a90")
	d.Key("1")
	d.Expect("reply (records answer)", "withdraw (cancels decision)")
	d.Key("1")
	d.Type("Yes use v2")
	d.Key("enter")
	if !f.called("@/r/tui decision answer tui-aaa111 " + testD2 + " --reply Yes use v2") {
		t.Fatalf("question reply argv wrong: %v", f.writes())
	}

	// Pick suggestion #2 -> withdraw.
	d.Key("enter")
	d.Key("2")
	d.Expect("accept (resolves decision; status unchanged)", "decline (resolves decision; status unchanged)", "withdraw (cancels decision)")
	d.Key("3")
	d.Type("Superseded")
	d.Key("enter")
	if !f.called("@/r/tui decision withdraw tui-aaa111 " + testD3 + " --reason Superseded") {
		t.Fatalf("suggestion withdraw argv wrong: %v", f.writes())
	}
}

func TestProjectViewPendingTabLoadsAndAnswers(t *testing.T) {
	f := projectFake()
	f.on("", "list --project tui --pending", goldenPendingFixture)
	env := testEnv(f)
	pv := newProjectView(env, "tui")
	d := drive(t, New(env, Options{Stack: []view{pv}}))
	d.Expect("tui-aaa111")
	// Shift+tab from Ready wraps cleanly to Pending tab!
	d.Key("shift+tab")
	d.Expect("pending decisions", "6f76bdab", "legacy:awaiting_user")
	if !f.called("list --project tui --pending") {
		t.Fatal("Pending tab must call list --project tui --pending")
	}
}

func TestPendingTargetsUseParkedWorktreeAndPendingViewScrolls(t *testing.T) {
	// 1. Standalone pendingView resolves parked worktree (/wt/c) before running decision answer.
	f := newFake()
	f.on("", "list --project tui --pending", `{
		"observed_at":"2026-11-24T17:07:06Z",
		"coverage":{"complete":true,"unavailable_projects":[]},
		"items":[{
			"type":"decision",
			"task":{"id":"tui-ccc333","title":"Parked one","status":"todo","project":"tui"},
			"opened_at":"2026-11-24T16:00:00Z","age_seconds":600,
			"decision":{
				"id":"`+testD2+`","type":"decision","at":"2026-11-24T16:00:00Z","by":"keith",
				"subject":{"scope":"`+testScope+`"},
				"data":{"kind":"question","origin":"scope","text":"Answer in worktree?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":"`+testScope+`","scope_changed_since_open":false,
				"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T17:07:06Z"}
			}
		}],
		"warnings":[]
	}`)
	todoStatus := "todo"
	f.on("", "list --project tui --parked", parkedJSON(tasksctl.ParkedRow{
		ID: "tui-ccc333", Title: "Parked one", Status: &todoStatus,
		Park: &tasksctl.ParkInfo{At: "2026-11-24T16:00:00Z", NextStep: "wait", WaitingOn: "user", Worktree: "/wt/c"},
	}))
	f.on("", "list --project tui --status doing", rowsJSON())
	wantReply := "decision answer tui-ccc333 " + testD2 + " --reply Done"
	f.on("/wt/c", wantReply, decAnswerOK)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newPendingView(env, "tui")}})
	d := drive(t, app)
	d.Expect("1b4e28ba")
	d.Key("enter")
	d.Key("1")
	d.Type("Done")
	d.Key("enter")
	if !f.called(fakeKey("/wt/c", wantReply)) {
		t.Fatalf("pendingView decision answer must run in parked worktree /wt/c, got %v", f.writes())
	}

	// 2. ProjectView Pending tab also resolves parked worktree (/wt/c) from prime.Parked.
	f2 := projectFake()
	f2.on("", "list --project tui --pending", `{
		"observed_at":"2026-11-24T17:07:06Z",
		"coverage":{"complete":true,"unavailable_projects":[]},
		"items":[{
			"type":"decision",
			"task":{"id":"tui-ccc333","title":"Parked one","status":"todo","project":"tui"},
			"opened_at":"2026-11-24T16:00:00Z","age_seconds":600,
			"decision":{
				"id":"`+testD2+`","type":"decision","at":"2026-11-24T16:00:00Z","by":"keith",
				"subject":{"scope":"`+testScope+`"},
				"data":{"kind":"question","origin":"scope","text":"Answer in tab?","source":"direct"},
				"state":"open","resolution":null,"seen_scope":"`+testScope+`","scope_changed_since_open":false,
				"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T17:07:06Z"}
			}
		}],
		"warnings":[]
	}`)
	f2.on("/wt/c", wantReply, decAnswerOK)
	env2 := testEnv(f2)
	pv := newProjectView(env2, "tui")
	d2 := drive(t, New(env2, Options{Stack: []view{pv}}))
	d2.Expect("tui-aaa111")
	d2.Key("shift+tab")
	d2.Expect("1b4e28ba")
	d2.Key("enter")
	d2.Key("1")
	d2.Type("Done")
	d2.Key("enter")
	if !f2.called(fakeKey("/wt/c", wantReply)) {
		t.Fatalf("projectView Pending tab must run decision answer in parked worktree /wt/c, got %v", f2.writes())
	}

	// 3. Standalone pendingView scrolls its viewport when selection moves past height.
	items := make([]tasksctl.PendingItem, 6)
	for i := range items {
		txt := "Item text"
		items[i] = tasksctl.PendingItem{
			Type:   "legacy_wait",
			Task:   tasksctl.PendingTaskRef{ID: "tui-00000" + string(rune('1'+i)), Title: "Row " + string(rune('1'+i)), Status: "idea", Project: "tui"},
			Source: "awaiting_user",
			Text:   &txt,
		}
	}
	vp := newPendingView(env, "")
	vp.data = &tasksctl.PendingResult{Coverage: tasksctl.PendingCoverage{Complete: true}, Items: items}
	vp.sel = 5
	rendered := vp.render(120, 4)
	if !strings.Contains(rendered, "▸") || !strings.Contains(rendered, "tui-000006") {
		t.Fatalf("pendingView must scroll to keep selected item 6 visible:\n%s", rendered)
	}
}
