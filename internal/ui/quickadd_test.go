package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"tasks-tui/internal/identity"
	"tasks-tui/internal/quickadd"
)

func TestQuickAddFilesIntoTheViewsProject(t *testing.T) {
	f := projectFake()
	f.on("", "add tint the strip --project tui --status idea -p 3 --size s --tag ui", `{"id":"tui-new001","action":"created","warnings":[]}`)
	f.on("", "tags --project tui", `{"tags":[],"warnings":["dictionary unreadable"]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Expect("tui-aaa111")
	d.Key("a")
	d.Expect("quick add", "→ into tui")
	d.Type("?tint the strip #ui !3 ~s")
	d.Expect(`add "tint the strip" --project tui --status idea -p 3 --size s --tag ui`)
	if !f.called("tags --project tui") || len(app.Messages()) != 0 {
		t.Fatalf("the tag lookup's warnings are dropped: %+v", app.Messages())
	}
	d.Key("enter")
	d.ExpectNot("quick add")
	if !logged(app, LevelInfo, "created tui-new001") {
		t.Fatalf("the filed id must be logged: %+v", app.Messages())
	}
	if !f.called("add tint the strip --project tui --status idea -p 3 --size s --tag ui") {
		t.Fatalf("add argv: %v", f.calls)
	}
	if n := countCalls(f, "ready --project tui"); n != 2 {
		t.Fatalf("the view must reload after add; ready loaded %d times", n)
	}
}

func TestQuickAddRefusesErrorsAndTargetsOtherProject(t *testing.T) {
	f := projectFake()
	f.on("", "add fix it --project ops", `{"id":"ops-new002","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("a")
	d.Type("fix it !9")
	d.Expect("!9: priority must be 0–4")
	d.Key("enter")
	d.Expect("quick add", "!9: priority must be 0–4")
	d.Key("backspace")
	d.Key("backspace")
	d.Type(">ops")
	d.Expect(`add "fix it" --project ops`)
	d.Key("enter")
	if !logged(app, LevelInfo, "created ops-new002") {
		t.Fatalf("the filed id must be logged: %+v", app.Messages())
	}
	if !f.called("add fix it --project ops") {
		t.Fatalf("add argv: %v", f.calls)
	}
	d.Key("a")
	d.Key("esc")
	d.ExpectNot("quick add")
	if countCalls(f, "add fix it --project ops") != 1 {
		t.Fatal("esc must not file anything")
	}
}

func TestQuickAddFromProjectsViewUsesHighlightedProject(t *testing.T) {
	f := newFake()
	f.on("", "projects", projectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project tui", primeJSON("tui"))
	f.on("", "add note --project tui", `{"id":"tui-new003","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectsView(env)}})
	d := drive(t, app)
	d.Key("a")
	d.Type("note")
	d.Key("enter")
	if !logged(app, LevelInfo, "created tui-new003") {
		t.Fatalf("the filed id must be logged: %+v", app.Messages())
	}
}

func TestQuickAddBodyIsVerbatim(t *testing.T) {
	f := projectFake()
	f.on("", "add x --project tui -b body  kept  ", `{"id":"tui-new004","action":"created","warnings":[]}`)
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("a")
	d.Type("x -- body  kept  ")
	d.Key("enter")
	if !f.called("add x --project tui -b body  kept  ") {
		t.Fatalf("body must reach add verbatim: %v", f.calls)
	}
}

func TestQuickAddUnreachableProjectRequiresExplicitProject(t *testing.T) {
	f := newFake()
	f.on("", "projects", projectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project tui", primeJSON("tui"))
	env := testEnv(f)
	pv := newProjectsView(env)
	app := New(env, Options{Stack: []view{pv}})
	d := drive(t, app)
	chord(t, d, app, "s", "p")
	d.Key("k")
	d.Key("a")
	d.Type("title")
	d.Expect("no project: add >prefix")
	d.Key("enter")
	d.Expect("quick add", "no project: add >prefix")
	if len(f.writes()) != 0 {
		t.Fatalf("unreachable project must not add: %v", f.writes())
	}
}

func TestStyledLineColoursTokens(t *testing.T) {
	s := NewStyles(identity.Tone{Dark: true, SatScale: 1})
	line := "fix #ui !2"
	spec, _ := quickadd.Parse(line, quickadd.Prefixes{"tui"}, "tui")
	out := styledLine(s, line, spec.Tokens, 2)
	if !strings.Contains(out, "#ui") || !strings.Contains(out, "!2") || !strings.Contains(out, "fix") {
		t.Fatalf("styled line lost text: %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("tokens must carry ANSI styling")
	}
}

func TestStyledPreviewColoursFlagsAndKeepsQuotedCommand(t *testing.T) {
	s := NewStyles(identity.Tone{Dark: true, SatScale: 1})
	spec, err := quickadd.Parse(`?fix this #ui !2 ~s ^mid @3d >tui -- body text`, quickadd.Prefixes{"tui"}, "tui")
	if err != nil {
		t.Fatal(err)
	}
	out := styledArgs(s, spec, 2)
	want := `add "fix this" --project tui --status idea -p 2 --size s --complexity mid --every 3d --tag ui -b "body text"`
	if got := ansi.Strip(out); got != want {
		t.Fatalf("preview\n got %q\nwant %q", got, want)
	}
	for _, colored := range []string{
		s.Accent(2).Render("--project tui"),
		s.PriorityStyle(2).Render("-p 2"),
		s.Muted.Render("--size s"),
		s.Periodic.Render("--every 3d"),
	} {
		if !strings.Contains(out, colored) {
			t.Fatalf("preview missing token-class styling %q in %q", colored, out)
		}
	}
}
