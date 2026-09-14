package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"tasks-tui/internal/config"
	"tasks-tui/internal/identity"
	"tasks-tui/internal/tasksctl"
	"tasks-tui/internal/ui"
)

var idRe = regexp.MustCompile(`^([a-z0-9]+)-([0-9a-f]{6})$`)

// resolveStart is spec §5 and §4.1: which views open for an argument, --all, or the
// cwd. Warnings from the lookups it makes are returned for the App's log (spec §11).
func resolveStart(ctx context.Context, env *ui.Env, arg string, all bool, cwd string) ([]ui.View, []string, error) {
	projects := ui.NewProjectsView(env)
	if all {
		return []ui.View{projects}, nil, nil
	}
	if arg == "" {
		if prefix := prefixOfCwd(env.Roots, cwd); prefix != "" {
			return []ui.View{projects, ui.NewProjectView(env, prefix)}, nil, nil
		}
		return []ui.View{projects}, nil, nil
	}
	if _, ok := env.Roots[arg]; ok {
		return []ui.View{projects, ui.NewProjectView(env, arg)}, nil, nil
	}
	m := idRe.FindStringSubmatch(arg)
	if m == nil {
		return nil, nil, fmt.Errorf("%q is neither a registered prefix nor a task id", arg)
	}
	prefix := m[1]
	root, ok := env.Roots[prefix]
	if !ok {
		return nil, nil, fmt.Errorf("%s: prefix %q is not registered", arg, prefix)
	}
	tgt := ui.Target{ID: arg, Prefix: prefix}
	var warnings []string
	parked, err := env.Client.Parked(ctx, prefix)
	if err != nil {
		return nil, nil, err
	}
	for _, w := range parked.Warnings {
		warnings = append(warnings, "list --project "+prefix+" --parked: "+w)
	}
	found := false
	for _, row := range parked.Tasks {
		if row.ID == arg {
			tgt.Title, tgt.Park, tgt.Claim, found = row.Title, row.Park, row.Claim, true
		}
	}
	if !found {
		res, err := env.Client.Show(ctx, root, arg)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", arg, err)
		}
		for _, w := range res.Warnings {
			warnings = append(warnings, "show "+arg+": "+w)
		}
		tgt.Title, tgt.Park, tgt.Claim = res.Task.Title, res.Park, res.Claim
	}
	return []ui.View{projects, ui.NewProjectView(env, prefix), ui.NewTaskView(env, tgt)}, warnings, nil
}

func resolveIdentity(projects []tasksctl.Project, cfg config.Identity, pins identity.Pins, git identity.GitRemote) (map[string]int, error) {
	slots := make(map[string]int, len(projects))
	for _, project := range projects {
		if slot, ok := cfg.Slot[project.Prefix]; ok {
			slots[project.Prefix] = slot
			continue
		}
		remote := git
		if !project.Reachable {
			remote = func(string) (string, error) { return "", nil }
		}
		slot, err := identity.SlotFor(project.Root, pins, remote)
		if err != nil {
			return nil, fmt.Errorf("identity for %s: %w", project.Prefix, err)
		}
		slots[project.Prefix] = slot
	}
	return slots, nil
}

func prefixOfCwd(roots map[string]string, cwd string) string {
	canon := func(p string) string {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			return r
		}
		return filepath.Clean(p)
	}
	c := canon(cwd)
	best, bestLen := "", 0
	for prefix, root := range roots {
		r := canon(root)
		if (c == r || strings.HasPrefix(c, r+string(filepath.Separator))) && len(r) > bestLen {
			best, bestLen = prefix, len(r)
		}
	}
	return best
}
