// Package launch opens a terminal window running a coding agent on a task (spec §7).
// It never changes the task's status; the launched agent runs `tasks start` itself.
package launch

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"syscall"
)

type Harness struct {
	Command []string          `toml:"command"`
	Env     map[string]string `toml:"env"`
}

type Config struct {
	Terminal []string           `toml:"terminal"`
	Prompt   string             `toml:"prompt"`
	Harness  map[string]Harness `toml:"harness"`
}

// Default is spec §7's table, verified against the installed binaries on 2026-09-13.
func Default() Config {
	return Config{
		Terminal: []string{"kitty", "--directory", "{dir}", "--"},
		Prompt:   "Run `tasks start {id}` and continue that task: {title}",
		Harness: map[string]Harness{
			"claude":   {Command: []string{"claude", "{prompt}"}},
			"codex":    {Command: []string{"codex", "{prompt}"}},
			"opencode": {Command: []string{"opencode", "--prompt", "{prompt}"}},
			"crush":    {Command: []string{"crush"}},
		},
	}
}

// Names lists the configured harnesses in a stable order for the picker.
func (c Config) Names() []string {
	names := make([]string, 0, len(c.Harness))
	for name := range c.Harness {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Plan is a fully substituted spawn: argv, directory, environment.
type Plan struct {
	Argv     []string
	Dir      string
	Env      []string
	NoPrompt bool // the harness command has no {prompt}; the status line says so
}

// Build substitutes {dir} in the terminal, {id}/{title} in the prompt, and {prompt}
// in the harness command. The environment is environ minus the TUI's session identity
// plus the harness env; agent variables are kept on purpose (spec §7).
func Build(cfg Config, harness, dir, id, title string, environ []string) (Plan, error) {
	h, ok := cfg.Harness[harness]
	if !ok {
		return Plan{}, fmt.Errorf("launch: no harness %q; configured: %s", harness, strings.Join(cfg.Names(), ", "))
	}
	if len(cfg.Terminal) == 0 {
		return Plan{}, fmt.Errorf("launch: terminal command is empty")
	}
	if len(h.Command) == 0 {
		return Plan{}, fmt.Errorf("launch: harness %q has an empty command", harness)
	}
	prompt := strings.NewReplacer("{id}", id, "{title}", title).Replace(cfg.Prompt)
	argv := make([]string, 0, len(cfg.Terminal)+len(h.Command))
	for _, arg := range cfg.Terminal {
		argv = append(argv, strings.ReplaceAll(arg, "{dir}", dir))
	}
	noPrompt := true
	for _, arg := range h.Command {
		if strings.Contains(arg, "{prompt}") {
			noPrompt = false
		}
		argv = append(argv, strings.ReplaceAll(arg, "{prompt}", prompt))
	}
	env := make([]string, 0, len(environ)+len(h.Env))
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		if name != "TASKS_SESSION" && name != "TASKS_SESSION_PID" {
			env = append(env, kv)
		}
	}
	keys := make([]string, 0, len(h.Env))
	for key := range h.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+h.Env[key])
	}
	return Plan{Argv: argv, Dir: dir, Env: env, NoPrompt: noPrompt}, nil
}

// Spawn starts the plan in its own session with stdio detached, and reaps it in the
// background. What the harness does afterwards is not the TUI's business.
func Spawn(p Plan) error {
	cmd := exec.Command(p.Argv[0], p.Argv[1:]...)
	cmd.Dir = p.Dir
	cmd.Env = p.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
