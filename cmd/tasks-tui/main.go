// Command tasks-tui is a keyboard-driven terminal front end for the tasks tracker.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/config"
	"tasks-tui/internal/identity"
	"tasks-tui/internal/launch"
	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/ui"
)

const version = "0.1.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "tasks-tui:", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("tasks-tui", flag.ContinueOnError)
	all := fs.Bool("all", false, "open the Projects view even inside a registered project")
	cfgPath := fs.String("config", "", "config file (default ~/.config/tasks-tui/config.toml)")
	showVersion := fs.Bool("version", false, "print the version")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: tasks-tui [--all] [--config path] [<prefix> | <task-id>]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	if *showVersion {
		fmt.Println("tasks-tui", version)
		return nil
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("at most one argument: a prefix or a task id")
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}
	client := &tasksctl.Client{R: tasksctl.ExecRunner{Bin: cfg.Tasks, Env: tasksctl.ChildEnv(os.Environ(), os.Getpid())}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	projects, err := client.Projects(ctx)
	if err != nil {
		return fmt.Errorf("%s projects: %w", cfg.Tasks, err)
	}

	familiar := filepath.Join(configHome(), "familiar")
	pins, err := identity.LoadPins(filepath.Join(familiar, "identities.yaml"))
	if err != nil {
		return err
	}
	tone, err := identity.LoadTone(filepath.Join(familiar, "scheme.json"))
	if err != nil {
		return err
	}
	slots, err := resolveIdentity(projects.Projects, cfg.Identity, pins, identity.GitOriginURL)
	if err != nil {
		return err
	}
	env := &ui.Env{
		Client:  client,
		Styles:  ui.NewStyles(tone),
		Config:  cfg,
		Slots:   slots,
		Roots:   map[string]string{},
		Exists:  func(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() },
		Environ: os.Environ(),
		Spawn:   launch.Spawn,
		Timeout: 10 * time.Second,
	}
	for _, p := range projects.Projects {
		env.Roots[p.Prefix] = p.Root
		env.Prefixes = append(env.Prefixes, p.Prefix)
	}

	cwd, _ := os.Getwd()
	stack, err := resolveStart(ctx, env, fs.Arg(0), *all, cwd)
	if err != nil {
		return err
	}
	app := ui.New(env, ui.Options{Refresh: time.Duration(cfg.RefreshSeconds) * time.Second, Stack: stack})
	_, err = tea.NewProgram(app).Run()
	return err
}

func configHome() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return x
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
