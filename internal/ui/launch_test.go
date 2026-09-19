package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/khughitt/tasks-tui/internal/launch"
)

func TestLaunchSpawnsInTheParkedWorktreeWithoutChangingStatus(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	var got launch.Plan
	env.Spawn = func(p launch.Plan) error { got = p; return nil }
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	d.Key("2")
	d.Expect("tui-ccc333")
	d.Key("j")
	chord(t, d, app, "c", "c")
	d.Expect("launch on tui-ccc333", "1 claude", "2 codex", "3 crush", "4 opencode")
	d.Key("q")
	if app.overlay != nil || d.Quit {
		t.Fatal("q closes the picker without quitting")
	}
	chord(t, d, app, "c", "c")
	d.Expect("launch on tui-ccc333", "─")
	d.Key("1")
	d.Expect("launched claude on tui-ccc333 in /wt/c")
	want := []string{"kitty", "--directory", "/wt/c", "--", "claude", "Run `tasks start tui-ccc333` and continue that task: parked one"}
	if !slices.Equal(got.Argv, want) || got.Dir != "/wt/c" {
		t.Fatalf("plan %+v", got)
	}
	if w := f.writes(); len(w) != 0 {
		t.Fatalf("launch must not run any task command: %v", w)
	}
}

func TestPickerHighlightsOnSurface(t *testing.T) {
	s := NewStyles(darkTone)
	o := &pickerOverlay{styles: s, slot: 2, title: "launch", items: []string{"claude", "codex"}}
	out := o.render(80)
	if !strings.Contains(out, s.Gutter(2).Render(" 1 claude ")) {
		t.Fatalf("selected item on the slot surface: %q", out)
	}
}

func TestLaunchWithoutPromptSaysSoAndSpawnErrorsShow(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	env.Spawn = func(launch.Plan) error { return nil }
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	chord(t, d, app, "c", "c")
	d.Key("j")
	d.Key("j")
	d.Key("enter") // claude, codex, crush, opencode
	d.Expect("launched crush on tui-aaa111 in /r/tui (no initial prompt; prompt copied)")
	env.Spawn = func(launch.Plan) error { return errors.New("kitty: not found") }
	chord(t, d, app, "c", "c")
	d.Key("esc")
	d.ExpectNot("launch on")
	chord(t, d, app, "c", "c")
	d.Key("1")
	d.Expect("launch: kitty: not found")
}

func TestNoPromptLaunchCopiesTheRenderedPrompt(t *testing.T) {
	env := testEnv(projectFake())
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	drive(t, app)
	app.openLaunch()
	msg := app.overlay.(*pickerOverlay).onPick("crush")().(launchMsg)
	_, cmd := app.Update(msg)
	if got := fmt.Sprint(cmd()); got != "Run `tasks start tui-aaa111` and continue that task: Title of tui-aaa111" {
		t.Fatalf("clipboard = %q", got)
	}
}

func TestLaunchNeedsATarget(t *testing.T) {
	f := newFake()
	f.on("", "projects", projectsJSON())
	f.on("", "list --all-projects --status doing", rowsJSON())
	f.on("", "list --all-projects --parked", parkedJSON())
	f.on("", "prime --project tui", primeJSON("tui"))
	env := testEnv(f)
	app := New(env, Options{Stack: []view{newProjectsView(env)}})
	d := drive(t, app)
	chord(t, d, app, "c", "c")
	d.Expect("no task highlighted")
}

func TestLaunchChordSpawnsAFixedHarnessOrSaysNotConfigured(t *testing.T) {
	f := projectFake()
	env := testEnv(f)
	var got launch.Plan
	env.Spawn = func(p launch.Plan) error { got = p; return nil }
	app := New(env, Options{Stack: []view{newProjectView(env, "tui")}})
	d := drive(t, app)
	chord(t, d, app, "c", "o")
	d.Expect("launched codex on tui-aaa111 in /r/tui")
	if app.overlay != nil || got.Argv[len(got.Argv)-2] != "codex" {
		t.Fatalf("c o spawns codex without the picker: %+v", got)
	}
	delete(env.Config.Launch.Harness, "crush")
	chord(t, d, app, "c", "r")
	d.Expect("launch: crush is not configured")
	chord(t, d, app, "c", "c")
	d.Expect("launch on tui-aaa111", "1 claude", "2 codex", "3 opencode")
}
