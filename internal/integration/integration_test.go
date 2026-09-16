// Package integration drives the TUI against the real tasks binary in a scratch
// project whose registry lives under a temporary XDG_CONFIG_HOME (spec §12).
package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tasks-tui/internal/config"
	"tasks-tui/internal/identity"
	"tasks-tui/internal/launch"
	"tasks-tui/internal/quickadd"
	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/ui"
	"tasks-tui/internal/uitest"
)

func TestProjectViewAgainstRealBinary(t *testing.T) {
	bin, err := exec.LookPath("tasks")
	if err != nil {
		t.Skip("tasks not on PATH")
	}
	// Isolate everything the binary writes outside the repo: the registry
	// (XDG_CONFIG_HOME) and the claim/park store (XDG_STATE_HOME). Drop every TASKS_*
	// variable the invoking shell may carry rather than blanking it: the CLI rejects an
	// empty TASKS_FORMAT.
	xdgConfig, xdgState := t.TempDir(), t.TempDir()
	repo := t.TempDir()
	var environ []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "TASKS_") || strings.HasPrefix(kv, "XDG_CONFIG_HOME=") || strings.HasPrefix(kv, "XDG_STATE_HOME=") {
			continue
		}
		environ = append(environ, kv)
	}
	environ = append(environ, "XDG_CONFIG_HOME="+xdgConfig, "XDG_STATE_HOME="+xdgState)
	sh := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repo
		cmd.Env = environ
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	sh("git", "init", "-q")
	sh(bin, "init", "--prefix", "zz")
	sh(bin, "add", "first task", "-p", "1", "--tag", "it")
	sh(bin, "add", "second task", "-p", "2")

	client := &tasksctl.Client{R: tasksctl.ExecRunner{Bin: bin, Env: tasksctl.ChildEnv(environ, os.Getpid())}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projects, err := client.Projects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	root := ""
	for _, p := range projects.Projects {
		if p.Prefix == "zz" {
			root = p.Root
		}
	}
	if root == "" {
		t.Fatalf("scratch project not registered: %+v", projects.Projects)
	}
	env := &ui.Env{Client: client, Styles: ui.NewStyles(identity.Tone{Dark: true, SatScale: 1}), Config: config.Default(),
		Slots: map[string]int{"zz": 0}, Roots: map[string]string{"zz": root}, Prefixes: quickadd.Prefixes{"zz"},
		Exists:  func(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() },
		Environ: environ, Spawn: func(launch.Plan) error { return nil }, Timeout: 10 * time.Second}

	app := ui.New(env, ui.Options{Stack: []ui.View{ui.NewProjectView(env, "zz")}})
	d := uitest.New(t, app, 120, 40)
	logged := func(text string) bool {
		for _, m := range app.Messages() {
			if strings.Contains(m.Text, text) {
				return true
			}
		}
		return false
	}
	d.Expect("first task", "second task", "P1", "[it]")
	d.Key("space") // start the highlighted (P1) task through the real binary
	if !logged("start zz-") {
		t.Fatalf("start must be logged: %+v", app.Messages())
	}
	d.Key("2") // the Doing tab shows it, claimed by this process's session
	d.Expect("first task", "doing", "◆")
	d.Key("p")
	d.Type("hand back")
	d.Key("ctrl+u")
	d.Key("enter")
	if !logged("park zz-") {
		t.Fatalf("park must be logged: %+v", app.Messages())
	}
	d.Key("2")
	d.Expect("⏸", "user", "first task")
	if strings.Count(d.Screen(), "first task") != 1 {
		t.Fatal("a parked doing task renders once")
	}
	d.Key("a")
	d.Type("third task #it !3")
	d.Key("enter")
	if !logged("created zz-") {
		t.Fatalf("add must be logged: %+v", app.Messages())
	}
	d.Key("1")
	d.Expect("third task")
	d.Key("q")
	if !d.Quit {
		t.Fatal("q must quit")
	}

	// Every change went through the binary, into the scratch project only.
	entries, _ := filepath.Glob(filepath.Join(repo, "tasks", "zz-*.md"))
	if len(entries) != 3 {
		t.Fatalf("expected three task files, found %v", entries)
	}
}
