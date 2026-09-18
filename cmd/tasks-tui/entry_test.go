package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"tasks-tui/internal/config"
	"tasks-tui/internal/identity"
	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/ui"
)

type replyRunner struct {
	replies map[string]string
	calls   []string
}

func (r *replyRunner) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	k := strings.Join(args, " ")
	if dir != "" {
		k = "@" + dir + " " + k
	}
	r.calls = append(r.calls, k)
	if rep, ok := r.replies[k]; ok {
		return []byte(rep), nil
	}
	return nil, &tasksctl.Error{Kind: "task_not_found", Detail: k}
}

func env(r *replyRunner) *ui.Env {
	return &ui.Env{Client: &tasksctl.Client{R: r}, Roots: map[string]string{"tui": "/r/tui"}, Slots: map[string]int{"tui": 0},
		Exists: func(p string) bool { return p == "/wt/x" || p == "/r/tui" }}
}

func TestResolveStartByIdPrefersParkedFeed(t *testing.T) {
	parked, _ := json.Marshal(map[string]any{"tasks": []tasksctl.ParkedRow{{ID: "tui-abc123", Title: "T", Tags: []string{}, Park: &tasksctl.ParkInfo{At: "t", Worktree: "/wt/x", WaitingOn: "user", NextStep: "n", Session: "s", Owner: "o", Host: "h"}}}, "warnings": []string{"store pruned"}})
	r := &replyRunner{replies: map[string]string{"list --project tui --parked": string(parked)}}
	stack, err := resolveStart(context.Background(), env(r), "tui-abc123", false, "/elsewhere")
	if err != nil || len(stack) != 3 {
		t.Fatalf("stack=%d err=%v", len(stack), err)
	}
	for _, c := range r.calls {
		if strings.HasPrefix(c, "@/r/tui show") {
			t.Fatal("a parked task must not be looked up with show at the root first")
		}
	}
}

func TestResolveStartByIdFallsBackToRootShow(t *testing.T) {
	empty := `{"tasks":[],"warnings":[]}`
	show := `{"task":{"id":"tui-abc123","title":"T","status":"todo","priority":2,"size":null,"complexity":null,"process":null,"parallel":false,"every":null,"owner":null,"created":"2026-09-13T00:00:00Z","updated":"2026-09-13T00:00:00Z","started":null,"completed":null,"last_done":null,"depends":[],"parent":null,"tags":[],"source":null,"model":null,"agent":null,"spec":null,"plan":null,"step":null,"body":"","notes":[]},"spec_path":null,"plan_path":null,"step_found":null,"depends_on":[],"parent":null,"children":[],"claim":null,"park":null,"escalation":null,"periodic":null,"warnings":[]}`
	r := &replyRunner{replies: map[string]string{"list --project tui --parked": empty, "@/r/tui show tui-abc123": show}}
	if _, err := resolveStart(context.Background(), env(r), "tui-abc123", false, "/x"); err != nil {
		t.Fatal(err)
	}
	r = &replyRunner{replies: map[string]string{"list --project tui --parked": empty}}
	if _, err := resolveStart(context.Background(), env(r), "tui-abc123", false, "/x"); err == nil {
		t.Fatal("an id nobody can find is an error")
	}
	if _, err := resolveStart(context.Background(), env(r), "zz-abc123", false, "/x"); err == nil {
		t.Fatal("an unregistered prefix is an error")
	}
}

func TestResolveStartPrefixCwdAndAll(t *testing.T) {
	r := &replyRunner{replies: map[string]string{}}
	if s, _ := resolveStart(context.Background(), env(r), "tui", false, "/x"); len(s) != 2 {
		t.Fatalf("prefix arg: %d views", len(s))
	}
	if s, _ := resolveStart(context.Background(), env(r), "", false, "/r/tui/sub/dir"); len(s) != 2 {
		t.Fatalf("cwd inside a root: %d views", len(s))
	}
	if s, _ := resolveStart(context.Background(), env(r), "", true, "/r/tui"); len(s) != 1 {
		t.Fatalf("--all: %d views", len(s))
	}
	if s, _ := resolveStart(context.Background(), env(r), "", false, "/x"); len(s) != 1 {
		t.Fatalf("outside every root: %d views", len(s))
	}
}

func TestResolveIdentitySkipsGitForUnreachableProject(t *testing.T) {
	projects := []tasksctl.Project{{Prefix: "gone", Root: "/gone", Reachable: false}}
	called := false
	slots, err := resolveIdentity(projects, config.Identity{Slot: map[string]int{}}, identity.Pins{}, func(string) (string, error) {
		called = true
		return "", errors.New("must not run")
	})
	if err != nil || called || slots["gone"] != identity.AutoSlot("/gone") {
		t.Fatalf("slots=%v called=%v err=%v", slots, called, err)
	}
}
