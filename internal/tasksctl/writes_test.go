package tasksctl

import (
	"context"
	"slices"
	"testing"
)

type argvRunner struct {
	calls [][]string
	dirs  []string
	reply string
}

func (a *argvRunner) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	a.calls = append(a.calls, args)
	a.dirs = append(a.dirs, dir)
	return []byte(a.reply), nil
}

func TestWritesBuildArgv(t *testing.T) {
	r := &argvRunner{reply: `{"id":"tui-1","warnings":["took over a stale claim"]}`}
	c := &Client{R: r}
	ctx := context.Background()

	if _, err := c.Start(ctx, "/wt", "tui-1", false); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Start(ctx, "/wt", "tui-1", true); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Park(ctx, "/wt", "tui-1", ParkSpec{NextStep: "next", WaitingOnUser: true, Reason: "review"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Park(ctx, "/wt", "tui-1", ParkSpec{NextStep: "next"}); err != nil {
		t.Fatal(err)
	}
	res, err := c.Done(ctx, "/wt", "tui-1", "landed")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Done(ctx, "/wt", "tui-1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Drop(ctx, "/wt", "tui-1", "why"); err != nil {
		t.Fatal(err)
	}

	want := [][]string{
		{"start", "tui-1"},
		{"start", "--force", "tui-1"},
		{"park", "tui-1", "next", "--waiting-on", "user", "--reason", "review"},
		{"park", "tui-1", "next"},
		{"done", "tui-1", "landed"},
		{"done", "tui-1"},
		{"drop", "tui-1", "why"},
	}
	for i := range want {
		if !slices.Equal(r.calls[i], want[i]) || r.dirs[i] != "/wt" {
			t.Fatalf("call %d: %v in %q", i, r.calls[i], r.dirs[i])
		}
	}
	if !slices.Equal(res.Warnings, []string{"took over a stale claim"}) {
		t.Fatalf("warnings %v", res.Warnings)
	}
}

func TestAddAndEditTaskFields(t *testing.T) {
	r := &argvRunner{reply: `{"id":"tui-2","action":"created","warnings":[]}`}
	c := &Client{R: r}
	fields := TaskFields{Title: "title", Body: "body\ntext", Tags: []string{"x", "y"}}
	res, err := c.AddTask(context.Background(), "tui", fields)
	if err != nil || res.ID != "tui-2" || res.Action != "created" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	r.reply = `{"id":"tui-2","warnings":[]}`
	if _, err := c.Edit(context.Background(), "/wt", "tui-2", fields); err != nil {
		t.Fatal(err)
	}
	if r.dirs[0] != "" || !slices.Equal(r.calls[0], []string{"add", "title", "--project", "tui", "-b", "body\ntext", "--tag", "x", "--tag", "y"}) {
		t.Fatalf("argv %v dir %q", r.calls[0], r.dirs[0])
	}
	if r.dirs[1] != "/wt" || !slices.Equal(r.calls[1], []string{"edit", "tui-2", "--title", "title", "-b", "body\ntext", "--no-tags", "--tag", "x", "--tag", "y"}) {
		t.Fatalf("argv %v dir %q", r.calls[1], r.dirs[1])
	}
}
