# tasks-tui

A keyboard-driven terminal front end for the `tasks` tracker, on Bubble Tea. It shells
out to the `tasks` binary and reads its JSON; it never touches `tasks/*.md` itself.

    just setup      # deps and git hooks
    just install    # go install ./cmd/tasks-tui
    tasks-tui       # Projects view, or the project you are standing in
    tasks-tui tui   # one project by prefix
    tasks-tui tui-d6e352

Keys: `?` legend · `a` quick add · `l` launch an agent · `s` start · `p` park · `d` done · `x` drop · `W` messages · `q` quit.

Config: `~/.config/tasks-tui/config.toml` (all optional; see `docs/specs/2026-09-13-tasks-tui-v1-design.md` §9).
Design: `docs/specs/`. Plans: `docs/plans/`.
