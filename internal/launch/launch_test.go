package launch

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

var environ = []string{"HOME=/h", "TASKS_SESSION=tasks-tui:1", "TASKS_SESSION_PID=1", "TASKS_AGENT=x/y", "TASKS_MODEL=m"}

func TestBuildDefaults(t *testing.T) {
	cfg := Default()
	p, err := Build(cfg, "claude", "/wt", "tui-1", "Do the thing", environ)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"kitty", "--directory", "/wt", "--", "claude", "Run `tasks start tui-1` and continue that task: Do the thing"}
	if !slices.Equal(p.Argv, want) {
		t.Fatalf("argv %q", p.Argv)
	}
	if p.Dir != "/wt" || p.NoPrompt {
		t.Fatalf("plan %+v", p)
	}
	for _, kv := range p.Env {
		if strings.HasPrefix(kv, "TASKS_SESSION") {
			t.Fatalf("TUI session leaked: %v", p.Env)
		}
	}
	for _, keep := range []string{"HOME=/h", "TASKS_AGENT=x/y", "TASKS_MODEL=m"} {
		if !slices.Contains(p.Env, keep) {
			t.Fatalf("agent env dropped: %v", p.Env)
		}
	}
	p, _ = Build(cfg, "opencode", "/wt", "tui-1", "T", environ)
	if !slices.Equal(p.Argv[4:], []string{"opencode", "--prompt", "Run `tasks start tui-1` and continue that task: T"}) {
		t.Fatalf("opencode argv %q", p.Argv)
	}
	p, _ = Build(cfg, "crush", "/wt", "tui-1", "T", environ)
	if !p.NoPrompt || !slices.Equal(p.Argv[4:], []string{"crush"}) {
		t.Fatalf("crush plan %+v", p)
	}
}

func TestBuildEnvAndErrors(t *testing.T) {
	cfg := Default()
	cfg.Harness["crush"] = Harness{Command: []string{"crush"}, Env: map[string]string{"TASKS_MAX_COMPLEXITY": "mid"}}
	p, err := Build(cfg, "crush", "/wt", "tui-1", "T", environ)
	if err != nil || !slices.Contains(p.Env, "TASKS_MAX_COMPLEXITY=mid") {
		t.Fatalf("env %v err=%v", p.Env, err)
	}
	if _, err := Build(cfg, "nope", "/wt", "tui-1", "T", environ); err == nil {
		t.Fatal("unknown harness must error")
	}
	cfg.Terminal = nil
	if _, err := Build(cfg, "claude", "/wt", "tui-1", "T", environ); err == nil {
		t.Fatal("empty terminal must error")
	}
	if got := Default().Names(); !slices.Equal(got, []string{"claude", "codex", "crush", "opencode"}) {
		t.Fatalf("names %v", got)
	}
}

func TestBuildHarnessEnvOverridesInheritedValue(t *testing.T) {
	cfg := Default()
	cfg.Harness["crush"] = Harness{Command: []string{"crush"}, Env: map[string]string{"TASKS_MAX_COMPLEXITY": "mid"}}
	p, err := Build(cfg, "crush", "/wt", "tui-1", "T", append(environ, "TASKS_MAX_COMPLEXITY=low"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("ignored")
	cmd.Env = p.Env
	if !slices.Contains(cmd.Environ(), "TASKS_MAX_COMPLEXITY=mid") || slices.Contains(cmd.Environ(), "TASKS_MAX_COMPLEXITY=low") {
		t.Fatalf("effective environment %v", cmd.Environ())
	}
}

func TestSpawnDetachesAndRunsInDir(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	p := Plan{Argv: []string{"sh", "-c", "pwd > ran"}, Dir: dir, Env: []string{"PATH=" + os.Getenv("PATH")}}
	if err := Spawn(p); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		if b, err := os.ReadFile(marker); err == nil {
			real, _ := filepath.EvalSymlinks(dir)
			if strings.TrimSpace(string(b)) != real {
				t.Fatalf("ran in %q, want %q", b, real)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("spawned command never ran")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSpawnErrorsOnMissingBinary(t *testing.T) {
	if err := Spawn(Plan{Argv: []string{filepath.Join(t.TempDir(), "missing")}, Dir: t.TempDir()}); err == nil {
		t.Fatal("missing binary must error")
	}
}
