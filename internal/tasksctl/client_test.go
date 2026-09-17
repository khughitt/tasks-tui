package tasksctl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type recorder struct {
	fixture string
	calls   [][]string
	dirs    []string
}

func (r *recorder) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, args)
	r.dirs = append(r.dirs, dir)
	return os.ReadFile(filepath.Join("testdata", r.fixture))
}

func TestProjects(t *testing.T) {
	r := &recorder{fixture: "projects.json"}
	res, err := (&Client{R: r}).Projects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.calls[0], []string{"projects"}) || r.dirs[0] != "" {
		t.Fatalf("argv %v dir %q", r.calls[0], r.dirs[0])
	}
	if len(res.Projects) == 0 || res.Projects[0].Prefix == "" || res.Projects[0].Root == "" {
		t.Fatalf("decoded %+v", res.Projects)
	}
	if res.Warnings == nil {
		t.Fatal("warnings must decode to a non-nil slice")
	}
}

func TestPrimeSplitsRowShapes(t *testing.T) {
	r := &recorder{fixture: "prime.json"}
	res, err := (&Client{R: r}).Prime(context.Background(), "tasks")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.calls[0], []string{"prime", "--project", "tasks"}) {
		t.Fatalf("argv %v", r.calls[0])
	}
	if res.Counts.Todo == 0 && res.Counts.Idea == 0 {
		t.Fatalf("counts %+v", res.Counts)
	}
	_ = res.Ready
	_ = res.Parked
}

func TestListRowsAndClaimLiveness(t *testing.T) {
	r := &recorder{fixture: "list_rows.json"}
	res, err := (&Client{R: r}).List(context.Background(), "tasks", ListOpts{Status: []string{"done"}, SortUpdated: true})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"list", "--project", "tasks", "--status", "done", "--sort", "updated"}
	if !slices.Equal(r.calls[0], want) {
		t.Fatalf("argv %v", r.calls[0])
	}
	if res.Tasks[0].Claim == nil || res.Tasks[0].Claim.Live || res.Tasks[1].Claim == nil || !res.Tasks[1].Claim.Live {
		t.Fatalf("liveness not decoded: %+v %+v", res.Tasks[0].Claim, res.Tasks[1].Claim)
	}
	if res.Tasks[0].Status == "" || res.Tasks[0].Updated == "" {
		t.Fatalf("required row fields empty: %+v", res.Tasks[0])
	}
}

func TestParkedRowsAllowUnresolved(t *testing.T) {
	r := &recorder{fixture: "list_parked.json"}
	res, err := (&Client{R: r}).Parked(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.calls[0], []string{"list", "--all-projects", "--parked"}) {
		t.Fatalf("argv %v", r.calls[0])
	}
	last := res.Tasks[len(res.Tasks)-1]
	if last.Status != nil || last.Priority != nil || last.ChildCount != nil || last.Park == nil {
		t.Fatalf("unresolved row decoded wrongly: %+v", last)
	}
	if !last.Unresolved() || res.Tasks[0].Unresolved() {
		t.Fatal("Unresolved() must follow status == nil")
	}
	r2 := &recorder{fixture: "list_parked.json"}
	if _, err := (&Client{R: r2}).Parked(context.Background(), "tui"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r2.calls[0], []string{"list", "--project", "tui", "--parked"}) {
		t.Fatalf("argv %v", r2.calls[0])
	}
}

func TestParkedRowRejectsNullPark(t *testing.T) {
	var row ParkedRow
	err := row.UnmarshalJSON([]byte(`{"id":"x-1","title":"t","status":null,"priority":null,"size":null,"complexity":null,"process":null,"owner":null,"updated":null,"tags":[],"parent":null,"child_count":null,"open_descendant_count":null,"claim":null,"park":null,"escalation":null,"phase":null}`))
	var taskErr *Error
	if !errors.As(err, &taskErr) || taskErr.Kind != "decode" {
		t.Fatalf("park:null must be a decode error: %v", err)
	}
}

func TestShow(t *testing.T) {
	r := &recorder{fixture: "show.json"}
	res, err := (&Client{R: r}).Show(context.Background(), "/repo", "tasks-7a7386")
	if err != nil {
		t.Fatal(err)
	}
	if r.dirs[0] != "/repo" || !slices.Equal(r.calls[0], []string{"show", "tasks-7a7386"}) {
		t.Fatalf("dir %q argv %v", r.dirs[0], r.calls[0])
	}
	if res.Task.ID != "tasks-7a7386" || res.PlanPath == nil || len(res.DependsOn) != 1 || !res.DependsOn[0].Resolved || res.Parent == nil {
		t.Fatalf("decoded %+v", res)
	}
	if len(res.Task.Notes) == 0 || res.Task.Notes[0].Text == "" {
		t.Fatalf("notes %+v", res.Task.Notes)
	}
}

func TestMissingRequiredFieldIsAnError(t *testing.T) {
	err := requireKeys([]byte(`{"tasks":[]}`), "tasks", "warnings")
	if err == nil || !strings.Contains(err.Error(), "warnings") {
		t.Fatalf("err=%v", err)
	}
	var row Row
	if err := row.UnmarshalJSON([]byte(`{"id":"x-1","title":"t"}`)); err == nil {
		t.Fatal("a row without status must not decode")
	}
	full := `{"id":"x-1","title":"t","status":%s,"priority":2,"size":null,"complexity":null,"process":null,"owner":null,"updated":"2026-09-13T00:00:00Z","tags":[],"parent":null,"child_count":0,"open_descendant_count":0,"claim":%s,"park":null,"periodic":null}`
	if err := row.UnmarshalJSON([]byte(fmt.Sprintf(full, `"todo"`, "null"))); err != nil {
		t.Fatalf("a complete row must decode: %v", err)
	}
	if err := row.UnmarshalJSON([]byte(fmt.Sprintf(full, "null", "null"))); err == nil {
		t.Fatal("null in a non-nullable field must not decode")
	}
	claim := `{"owner":"o","session":"s","host":"h","pid":null,"worktree":"/w","started":"t","seen":"t"}`
	if err := row.UnmarshalJSON([]byte(fmt.Sprintf(full, `"todo"`, claim))); err == nil || !strings.Contains(err.Error(), "live") {
		t.Fatalf("a claim without live must not decode: %v", err)
	}
	var proj Project
	if err := proj.UnmarshalJSON([]byte(`{"prefix":"x"}`)); err == nil {
		t.Fatal("a project with only a prefix must not decode")
	}
	unreachable := `{"prefix":"x","root":"/x","reachable":false,"counts":null,"total":null,"last_activity":null}`
	if err := proj.UnmarshalJSON([]byte(unreachable)); err != nil || proj.Counts != nil {
		t.Fatalf("an unreachable project has null counts and total: %v", err)
	}
	var rel Relation
	if err := rel.UnmarshalJSON([]byte(`{"id":"x-1","title":null,"status":null,"resolved":false}`)); err == nil {
		t.Fatal("parent/children relations are strict")
	}
	var dep Dependency
	if err := dep.UnmarshalJSON([]byte(`{"id":"x-1","title":null,"status":null,"resolved":false}`)); err != nil || dep.Resolved {
		t.Fatalf("an unresolved dependency decodes with nulls: %v", err)
	}
}

func TestRoot(t *testing.T) {
	lit := literalRunner(`{"prefix":"tasks","root":"/r","warnings":[]}`)
	res, err := (&Client{R: lit}).Root(context.Background(), "tasks-1")
	if err != nil || res.Root != "/r" || res.Prefix != "tasks" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

type literalRunner string

func (l literalRunner) Run(context.Context, string, ...string) ([]byte, error) { return []byte(l), nil }

func TestSparseTaskJSON(t *testing.T) {
	ctx := context.Background()
	list := literalRunner(`{"tasks":[{"id":"sci-123456","title":"Sparse","status":"todo","priority":0,"updated":"2026-09-17T00:00:00Z","child_count":0,"open_descendant_count":0}],"warnings":[]}`)
	rows, err := (&Client{R: list}).List(ctx, "sci", ListOpts{})
	if err != nil {
		t.Fatal(err)
	}
	row := rows.Tasks[0]
	if row.Size != nil || row.Claim != nil || len(row.Tags) != 0 || row.Priority != 0 {
		t.Fatalf("sparse row: %+v", row)
	}
	parked := literalRunner(`{"tasks":[{"id":"sci-123456","title":"Sparse","parallel":false,"park":{"at":"t","next_step":"resume","waiting_on":"agent","session":"s","owner":"o","host":"h","worktree":"/w"}}],"warnings":[]}`)
	parks, err := (&Client{R: parked}).Parked(ctx, "sci")
	if err != nil {
		t.Fatal(err)
	}
	if !parks.Tasks[0].Unresolved() || parks.Tasks[0].Park.NextStep != "resume" || parks.Tasks[0].Park.Reason != nil {
		t.Fatalf("sparse park: %+v", parks.Tasks[0])
	}
	show := literalRunner(`{"task":{"id":"sci-123456","title":"Sparse","status":"todo","priority":0,"parallel":false,"created":"t","updated":"t","body":""},"warnings":[]}`)
	detail, err := (&Client{R: show}).Show(ctx, "/w", "sci-123456")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Task.Process != nil || len(detail.Task.Notes) != 0 || len(detail.Task.Depends) != 0 || len(detail.Children) != 0 || len(detail.DependsOn) != 0 || detail.Claim != nil {
		t.Fatalf("sparse detail: %+v", detail)
	}
	var periodic Periodic
	if err := periodic.UnmarshalJSON([]byte(`{"every":"30d","due_now":false}`)); err != nil || periodic.Due != nil || periodic.DueNow {
		t.Fatalf("sparse periodic: %+v, %v", periodic, err)
	}
	var claim ClaimInfo
	if err := claim.UnmarshalJSON([]byte(`{"owner":"o","session":"s","host":"h","worktree":"/w","started":"t","seen":"t","live":false}`)); err != nil || claim.PID != nil || claim.Live {
		t.Fatalf("sparse claim: %+v, %v", claim, err)
	}
	var dep Dependency
	if err := dep.UnmarshalJSON([]byte(`{"id":"sci-654321","resolved":false}`)); err != nil || dep.Resolved || dep.Title != nil {
		t.Fatalf("sparse dependency: %+v, %v", dep, err)
	}
}
