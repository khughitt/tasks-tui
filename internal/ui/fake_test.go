package ui

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"tasks-tui/internal/config"
	"tasks-tui/internal/identity"
	"tasks-tui/internal/launch"
	"tasks-tui/internal/quickadd"
	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/uitest"
)

type fakeRunner struct {
	mu      sync.Mutex
	replies map[string]string
	errors  map[string]*tasksctl.Error
	calls   []string
}

var darkTone = identity.Tone{Dark: true, SatScale: 1}

//lint:ignore U1000 shared by subsequent UI tests
func col(t *testing.T, line, tok string) int {
	t.Helper()
	i := strings.Index(line, tok)
	if i < 0 {
		t.Fatalf("no %q in %q", tok, line)
	}
	return lipgloss.Width(line[:i])
}

func newFake() *fakeRunner {
	return &fakeRunner{replies: map[string]string{}, errors: map[string]*tasksctl.Error{}}
}
func fakeKey(dir, args string) string {
	if dir != "" {
		return "@" + dir + " " + args
	}
	return args
}
func (f *fakeRunner) on(dir, args string, reply any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fakeKey(dir, args)
	switch reply := reply.(type) {
	case string:
		f.replies[key] = reply
	case *tasksctl.Error:
		f.errors[key] = reply
	default:
		data, _ := json.Marshal(reply)
		f.replies[key] = string(data)
	}
}
func (f *fakeRunner) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fakeKey(dir, strings.Join(args, " "))
	f.calls = append(f.calls, key)
	if err, ok := f.errors[key]; ok {
		return nil, err
	}
	if reply, ok := f.replies[key]; ok {
		return []byte(reply), nil
	}
	return nil, &tasksctl.Error{Kind: "fake", Detail: "no reply for " + key}
}
func (f *fakeRunner) called(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, call := range f.calls {
		if call == key {
			return true
		}
	}
	return false
}
func (f *fakeRunner) writes() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, call := range f.calls {
		if strings.HasPrefix(call, "@") {
			out = append(out, call)
		}
	}
	return out
}
func row(id, status string, priority int) tasksctl.Row {
	return tasksctl.Row{ID: id, Title: "Title of " + id, Status: status, Priority: priority, Updated: "2026-09-13T10:00:00Z", Tags: []string{"t"}}
}
func rowJSON(row tasksctl.Row) string { data, _ := json.Marshal(row); return string(data) }
func parkedRow(id, title, status, worktree string) tasksctl.ParkedRow {
	priority, zero := 2, 0
	updated := "2026-09-13T10:00:00Z"
	return tasksctl.ParkedRow{ID: id, Title: title, Status: &status, Priority: &priority, Updated: &updated, Tags: []string{}, ChildCount: &zero, OpenDescendantCount: &zero, Park: &tasksctl.ParkInfo{At: updated, WaitingOn: "user", NextStep: "n", Session: "sid:1", Owner: "o", Host: "h", Worktree: worktree}}
}
func rowsJSON(rows ...tasksctl.Row) string {
	if rows == nil {
		rows = []tasksctl.Row{}
	}
	data, _ := json.Marshal(map[string]any{"tasks": rows, "warnings": []string{}})
	return string(data)
}
func parkedJSON(rows ...tasksctl.ParkedRow) string {
	if rows == nil {
		rows = []tasksctl.ParkedRow{}
	}
	data, _ := json.Marshal(map[string]any{"tasks": rows, "warnings": []string{}})
	return string(data)
}
func testEnv(f *fakeRunner) *Env {
	styles := NewStyles(identity.Tone{Dark: true, SatScale: 1})
	styles.cursorBlink = false
	return &Env{Client: &tasksctl.Client{R: f}, Styles: styles, Config: config.Default(), Slots: map[string]int{"tui": 2, "ops": 1}, Roots: map[string]string{"tui": "/r/tui", "ops": "/r/ops"}, Prefixes: quickadd.Prefixes{"tui", "ops"}, Exists: func(path string) bool { return strings.HasPrefix(path, "/r/") || strings.HasPrefix(path, "/wt/") }, Environ: []string{"HOME=/h"}, Spawn: func(launch.Plan) error { return nil }, Timeout: 2 * time.Second}
}
func drive(t *testing.T, app *App) *uitest.Driver { t.Helper(); return uitest.New(t, app, 120, 40) }
