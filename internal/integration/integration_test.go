// Package integration drives the TUI against the real tasks binary in a scratch
// project whose registry lives under a temporary XDG_CONFIG_HOME.
package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/khughitt/tasks-tui/internal/config"
	"github.com/khughitt/tasks-tui/internal/launch"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
	"github.com/khughitt/tasks-tui/internal/ui"
	"github.com/khughitt/tasks-tui/internal/uitest"
)

func TestProjectViewAgainstRealBinary(t *testing.T) {
	bin, err := exec.LookPath("tasks")
	if err != nil {
		t.Skip("tasks not on PATH")
	}
	// Every subprocess gets an owned home, isolated registry/store and explicit Git config.
	xdgConfig, xdgState, home := t.TempDir(), t.TempDir(), t.TempDir()
	repo := t.TempDir()
	environ := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "XDG_CONFIG_HOME=" + xdgConfig,
		"XDG_STATE_HOME=" + xdgState, "USER=fixture", "LANG=C.UTF-8", "GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid"}
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
	realRepo, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"--show-toplevel", "--absolute-git-dir"} {
		cmd := exec.Command("git", "rev-parse", query)
		cmd.Dir = repo
		cmd.Env = environ
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		actual, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
		if err != nil {
			t.Fatal(err)
		}
		if query == "--show-toplevel" && actual != realRepo || query == "--absolute-git-dir" && !strings.HasPrefix(actual, realRepo+string(os.PathSeparator)) {
			t.Fatalf("Git escaped fixture: %s %s", query, actual)
		}
	}
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
	env := &ui.Env{Client: client, Styles: ui.NewStyles(), Config: config.Default(),
		Roots:   map[string]string{"zz": root},
		Exists:  func(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() },
		Environ: environ, Spawn: func(launch.Plan) error { return nil }, Timeout: 10 * time.Second}

	env.SetProjects(projects.Projects)
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
	d.Type("third task")
	d.Key("tab")
	d.Key("tab")
	d.Type("it")
	d.Key("ctrl+s")
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
