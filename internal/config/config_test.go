package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func write(t *testing.T, text string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMissingFileIsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "none.toml"))
	if err != nil || cfg.Tasks != "tasks" || cfg.RefreshSeconds != 30 || len(cfg.Launch.Harness) != 4 {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}

func TestOverridesMergeIntoDefaults(t *testing.T) {
	cfg, err := Load(write(t, `
tasks = "/opt/tasks"
refresh_seconds = 0
[launch]
terminal = ["ghostty", "--working-directory={dir}", "-e"]
[launch.harness.crush]
command = ["crush"]
env = { TASKS_MAX_COMPLEXITY = "mid" }
[identity]
slot = { tui = 2 }
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tasks != "/opt/tasks" || cfg.RefreshSeconds != 0 {
		t.Fatalf("scalars %+v", cfg)
	}
	if !slices.Equal(cfg.Launch.Terminal, []string{"ghostty", "--working-directory={dir}", "-e"}) {
		t.Fatalf("terminal %v", cfg.Launch.Terminal)
	}
	if cfg.Launch.Harness["crush"].Env["TASKS_MAX_COMPLEXITY"] != "mid" || len(cfg.Launch.Harness["claude"].Command) == 0 {
		t.Fatalf("harness merge %+v", cfg.Launch.Harness)
	}
	if cfg.Identity.Slot["tui"] != 2 {
		t.Fatalf("identity %+v", cfg.Identity)
	}
}

func TestHarnessEnvOverrideKeepsDefaultCommand(t *testing.T) {
	cfg, err := Load(write(t, `
[launch.harness.crush]
env = { TASKS_MAX_COMPLEXITY = "mid" }
`))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.Launch.Harness["crush"].Command, []string{"crush"}) {
		t.Fatalf("command %v", cfg.Launch.Harness["crush"].Command)
	}
}

func TestUnknownKeyIsAnError(t *testing.T) {
	_, err := Load(write(t, "refresh = 5\n"))
	if err == nil || !strings.Contains(err.Error(), "refresh") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidation(t *testing.T) {
	for _, text := range []string{
		"refresh_seconds = -1\n",
		"tasks = \"\"\n",
		"[identity]\nslot = { tui = 12 }\n",
		"[launch]\nterminal = []\n",
		"[launch.harness.crush]\ncommand = []\n",
	} {
		if _, err := Load(write(t, text)); err == nil {
			t.Errorf("%q must fail validation", text)
		}
	}
}

func TestDefaultPathHonoursXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	if got := DefaultPath(); got != "/xdg/tasks-tui/config.toml" {
		t.Fatalf("path %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	if !strings.HasSuffix(DefaultPath(), "/.config/tasks-tui/config.toml") {
		t.Fatalf("path %q", DefaultPath())
	}
}
