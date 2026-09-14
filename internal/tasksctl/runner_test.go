package tasksctl

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestChildEnvScrubsAndSetsSession(t *testing.T) {
	in := []string{
		"HOME=/home/x", "TASKS_FORMAT=pretty", "TASKS_COLOR=always", "TASKS_AGENT=claude-code/opus",
		"TASKS_MODEL=opus", "TASKS_MAX_COMPLEXITY=mid", "TASKS_SESSION=old", "TASKS_SESSION_PID=1",
		"PATH=/bin",
	}
	got := ChildEnv(in, 4242)
	for _, banned := range []string{"TASKS_FORMAT=", "TASKS_COLOR=", "TASKS_AGENT=", "TASKS_MODEL=", "TASKS_MAX_COMPLEXITY="} {
		for _, kv := range got {
			if strings.HasPrefix(kv, banned) {
				t.Fatalf("%q survived scrubbing: %v", banned, got)
			}
		}
	}
	for _, want := range []string{"HOME=/home/x", "PATH=/bin", "TASKS_SESSION=tasks-tui:4242", "TASKS_SESSION_PID=4242"} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
	if n := countPrefix(got, "TASKS_SESSION="); n != 1 {
		t.Fatalf("TASKS_SESSION set %d times", n)
	}
}

func countPrefix(env []string, p string) int {
	n := 0
	for _, kv := range env {
		if strings.HasPrefix(kv, p) {
			n++
		}
	}
	return n
}

func TestDecodeFailureEnvelope(t *testing.T) {
	err := decodeFailure([]byte(`{"error":{"detail":"task x not found","kind":"task_not_found"}}` + "\n"))
	if err.Kind != "task_not_found" || err.Detail != "task x not found" {
		t.Fatalf("got %+v", err)
	}
	if err.Error() != "task_not_found: task x not found" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestDecodeFailureNonJSON(t *testing.T) {
	err := decodeFailure([]byte("error: unexpected argument '-n' found\n"))
	if err.Kind != "unparseable" || !strings.Contains(err.Detail, "unexpected argument") {
		t.Fatalf("got %+v", err)
	}
}

func TestCommandShape(t *testing.T) {
	r := ExecRunner{Bin: "/usr/bin/tasks", Env: []string{"A=1"}}
	cmd := r.command(context.Background(), "/repo", "show", "x-1")
	if got := cmd.Args; !slices.Equal(got, []string{"/usr/bin/tasks", "-C", "/repo", "show", "x-1"}) {
		t.Fatalf("args %v", got)
	}
	if !slices.Equal(cmd.Env, []string{"A=1"}) {
		t.Fatalf("env %v", cmd.Env)
	}
	cmd = r.command(context.Background(), "", "projects")
	if got := cmd.Args; !slices.Equal(got, []string{"/usr/bin/tasks", "projects"}) {
		t.Fatalf("args without dir %v", got)
	}
}

func fakeTasks(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks")
	script := "#!/bin/sh\nif [ \"$1\" = fail ]; then printf '%s' '{\"error\":{\"kind\":\"claimed\",\"detail\":\"held\"}}' >&2; exit 3; fi\nprintf '%s' '{\"ok\":true}'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunSuccessAndFailure(t *testing.T) {
	r := ExecRunner{Bin: fakeTasks(t), Env: os.Environ()}
	out, err := r.Run(context.Background(), "", "projects")
	if err != nil || string(out) != `{"ok":true}` {
		t.Fatalf("out=%q err=%v", out, err)
	}
	_, err = r.Run(context.Background(), "", "fail")
	var e *Error
	if !errors.As(err, &e) || e.Kind != "claimed" || e.Detail != "held" {
		t.Fatalf("err=%v", err)
	}
	_, err = ExecRunner{Bin: filepath.Join(t.TempDir(), "missing")}.Run(context.Background(), "", "projects")
	if !errors.As(err, &e) || e.Kind != "spawn" {
		t.Fatalf("missing binary err=%v", err)
	}
	if _, ok := err.(*exec.Error); ok {
		t.Fatal("spawn failure must be wrapped as *Error")
	}
}
