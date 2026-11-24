package tasksctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestProjectColorContract(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		color        any
		index        any
		invalid      bool
	}{
		{name: "missing"},
		{name: "null", fields: `,"color":null,"color_256":null`},
		{name: "mixed hex", fields: `,"color":"#123AbC","color_256":42`, color: "#123AbC", index: float64(42)},
		{name: "rgb only", fields: `,"color":"#abcdef"`, color: "#abcdef"},
		{name: "low index", fields: `,"color":"#abcdef","color_256":16`, color: "#abcdef", index: float64(16)},
		{name: "high index", fields: `,"color":"#abcdef","color_256":255`, color: "#abcdef", index: float64(255)},
		{name: "unrelated field", fields: `,"unknown":"fine"`},
		{name: "orphan", fields: `,"color_256":42`, invalid: true},
		{name: "null rgb with index", fields: `,"color":null,"color_256":42`, invalid: true},
		{name: "empty", fields: `,"color":""`, invalid: true},
		{name: "short", fields: `,"color":"#abc"`, invalid: true},
		{name: "nonhex", fields: `,"color":"#12ggff"`, invalid: true},
		{name: "rgb type", fields: `,"color":42`, invalid: true},
		{name: "index type", fields: `,"color":"#abcdef","color_256":"42"`, invalid: true},
		{name: "fraction", fields: `,"color":"#abcdef","color_256":42.5`, invalid: true},
		{name: "index too low", fields: `,"color":"#abcdef","color_256":15`, invalid: true},
		{name: "index too high", fields: `,"color":"#abcdef","color_256":256`, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"prefix":"tui","root":"/r/tui","reachable":true,"counts":null,"total":null,"last_activity":null` + tc.fields + `}`
			var p Project
			err := json.Unmarshal([]byte(raw), &p)
			if tc.invalid {
				var typed *Error
				if !errors.As(err, &typed) || typed.Kind != "decode" {
					t.Fatalf("expected typed decode error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]any
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if fields["color"] != tc.color || fields["color_256"] != tc.index {
				t.Fatalf("color metadata lost: %s", encoded)
			}
		})
	}
}

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
	for _, p := range res.Projects {
		if p.Color != nil && p.Color256 != nil {
			return
		}
	}
	t.Fatal("captured projects must exercise authored color and its stand-in")
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

func TestActionExplanation(t *testing.T) {
	runner := &recorder{fixture: "show-actions.json"}
	result, err := (&Client{R: runner}).Show(context.Background(), "/repo", "alpha-000001")
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile("testdata/show-actions.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]any
	if err := json.Unmarshal(original, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got["actions"], want["actions"]) {
		t.Fatalf("decoded actions = %v, want producer actions %v", got["actions"], want["actions"])
	}
	if len(runner.calls) != 1 || !slices.Equal(runner.calls[0], []string{"show", "alpha-000001"}) {
		t.Fatalf("calls = %v", runner.calls)
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

func TestOpenListRetainsShelfCheckpointFields(t *testing.T) {
	r := &recorder{fixture: "list_rows.json"}
	res, err := (&Client{R: r}).List(context.Background(), "tui", ListOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.calls[0], []string{"list", "--project", "tui"}) {
		t.Fatalf("argv %v", r.calls[0])
	}
	shelf := res.Tasks[len(res.Tasks)-1]
	if shelf.ID != "tui-abc999" || shelf.Status != "shelved" {
		t.Fatalf("shelf lost: %+v", shelf)
	}
}

type captureRunner struct {
	reply string
	calls [][]string
	dirs  []string
}

func (c *captureRunner) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	c.dirs = append(c.dirs, dir)
	c.calls = append(c.calls, append([]string{}, args...))
	return []byte(c.reply), nil
}

func TestParkStageBArgvAndShowDecisionsDecode(t *testing.T) {
	ctx := context.Background()
	cap := &captureRunner{reply: `{"id":"tui-111111","warnings":[]}`}
	cli := &Client{R: cap}

	// Agent park and non-request user park keep exact existing argv.
	if _, err := cli.Park(ctx, "/wt", "tui-111111", ParkSpec{NextStep: "continue"}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cap.calls[0], []string{"park", "tui-111111", "continue"}) {
		t.Fatalf("agent park argv = %v", cap.calls[0])
	}
	if _, err := cli.Park(ctx, "/wt", "tui-111111", ParkSpec{NextStep: "fix env", WaitingOnUser: true, Reason: "environment"}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cap.calls[1], []string{"park", "tui-111111", "fix env", "--waiting-on", "user", "--reason", "environment"}) {
		t.Fatalf("non-request park argv = %v", cap.calls[1])
	}

	// Link-only human-request park passes --decision.
	const d1 = "6f76bdab-e417-4ae7-81ce-09fc70d6487a"
	if _, err := cli.Park(ctx, "/wt", "tui-111111", ParkSpec{NextStep: "review", WaitingOnUser: true, Reason: "review", DecisionID: d1}); err != nil {
		t.Fatal(err)
	}
	wantLink := []string{"park", "tui-111111", "review", "--waiting-on", "user", "--reason", "review", "--decision", d1}
	if !slices.Equal(cap.calls[2], wantLink) {
		t.Fatalf("link park argv = %v, want %v", cap.calls[2], wantLink)
	}

	// Inline-open commit-pinned approval park passes all Stage B flags.
	const commit = "0123456789abcdef0123456789abcdef01234567"
	if _, err := cli.Park(ctx, "/wt", "tui-111111", ParkSpec{
		NextStep:       "land after approval",
		WaitingOnUser:  true,
		Reason:         "approval",
		DecisionID:     d1,
		DecisionKind:   "approval",
		DecisionOrigin: "tasks-tui",
		DecisionText:   "Approve plan?",
		ArtifactRepo:   "records",
		ArtifactPath:   "tasks/docs/plans/p.md",
		ArtifactCommit: commit,
	}); err != nil {
		t.Fatal(err)
	}
	wantInline := []string{
		"park", "tui-111111", "land after approval",
		"--waiting-on", "user", "--reason", "approval",
		"--decision", d1, "--decision-kind", "approval", "--decision-origin", "tasks-tui", "--decision-text", "Approve plan?",
		"--artifact-repo", "records", "--artifact-path", "tasks/docs/plans/p.md", "--artifact-commit", commit,
	}
	if !slices.Equal(cap.calls[3], wantInline) {
		t.Fatalf("inline approval argv = %v, want %v", cap.calls[3], wantInline)
	}

	// Decode show JSON carrying decisions[] and park.decision_id.
	showRaw := `{
		"task":{"id":"tui-111111","title":"T","status":"todo","priority":2,"parallel":false,"created":"2026-11-24T00:00:00Z","updated":"2026-11-24T00:00:00Z","body":""},
		"park":{"at":"2026-11-24T00:00:00Z","next_step":"review","waiting_on":"user","reason":"review","decision_id":"6f76bdab-e417-4ae7-81ce-09fc70d6487a","session":"s","owner":"o","host":"h","worktree":"/wt"},
		"decisions":[{
			"id":"6f76bdab-e417-4ae7-81ce-09fc70d6487a","type":"decision","at":"2026-11-24T00:00:00Z","by":"keith",
			"subject":{"scope":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			"data":{"kind":"question","origin":"scope","text":"Which layout?"},
			"state":"open","resolution":null,"seen_scope":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","scope_changed_since_open":false,
			"open_dependents":{"tasks":[],"complete":true,"unavailable_projects":[],"observed_at":"2026-11-24T00:00:00Z"}
		}],
		"warnings":[]
	}`
	showRes, err := (&Client{R: literalRunner(showRaw)}).Show(ctx, "/wt", "tui-111111")
	if err != nil {
		t.Fatal(err)
	}
	if showRes.Park == nil || showRes.Park.DecisionID == nil || *showRes.Park.DecisionID != d1 {
		t.Fatalf("park.decision_id not decoded: %+v", showRes.Park)
	}
	if len(showRes.Decisions) != 1 || showRes.Decisions[0].ID != d1 || showRes.Decisions[0].Data.Kind != "question" {
		t.Fatalf("decisions not decoded: %+v", showRes.Decisions)
	}
}
