// Package tasksctl is the only place tasks-tui runs the tasks binary. Every call goes
// through a Runner; the typed Client on top builds argv and decodes JSON. The binary is
// the sole reader and writer of task data (spec §2).
package tasksctl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// Runner executes `tasks [-C dir] args...` and returns stdout on exit 0. A non-zero
// exit yields *Error decoded from stderr (spec §4). dir == "" omits -C.
type Runner interface {
	Run(ctx context.Context, dir string, args ...string) ([]byte, error)
}

// Error is the CLI's error envelope, or a failure to reach the CLI at all.
// Kind is what callers branch on ("claimed", "task_not_found", "spawn", "unparseable").
type Error struct {
	Kind   string
	Detail string
}

func (e *Error) Error() string { return e.Kind + ": " + e.Detail }

var scrubbed = []string{"TASKS_FORMAT", "TASKS_COLOR", "TASKS_AGENT", "TASKS_MODEL", "TASKS_MAX_COMPLEXITY", "TASKS_SESSION", "TASKS_SESSION_PID"}

// ChildEnv is environ without the scrubbed variables, plus the TUI's session identity.
func ChildEnv(environ []string, pid int) []string {
	out := make([]string, 0, len(environ)+2)
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		if !slices.Contains(scrubbed, name) {
			out = append(out, kv)
		}
	}
	return append(out, "TASKS_SESSION=tasks-tui:"+strconv.Itoa(pid), "TASKS_SESSION_PID="+strconv.Itoa(pid))
}

// ExecRunner runs the real binary with a fixed environment.
type ExecRunner struct {
	Bin string
	Env []string
}

// NewExecRunner returns a runner for bin with the TUI child environment.
func NewExecRunner(bin string) ExecRunner {
	return ExecRunner{Bin: bin, Env: ChildEnv(os.Environ(), os.Getpid())}
}

func (r ExecRunner) command(ctx context.Context, dir string, args ...string) *exec.Cmd {
	argv := make([]string, 0, len(args)+2)
	if dir != "" {
		argv = append(argv, "-C", dir)
	}
	argv = append(argv, args...)
	cmd := exec.CommandContext(ctx, r.Bin, argv...)
	cmd.Env = r.Env
	return cmd
}

func (r ExecRunner) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := r.command(ctx, dir, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, decodeFailure(stderr.Bytes())
		}
		return nil, &Error{Kind: "spawn", Detail: fmt.Sprintf("%s: %v", r.Bin, err)}
	}
	return stdout.Bytes(), nil
}

func decodeFailure(stderr []byte) *Error {
	var env struct {
		Error struct {
			Kind   string `json:"kind"`
			Detail string `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stderr), &env); err == nil && env.Error.Kind != "" {
		return &Error{Kind: env.Error.Kind, Detail: env.Error.Detail}
	}
	return &Error{Kind: "unparseable", Detail: strings.TrimSpace(string(stderr))}
}
