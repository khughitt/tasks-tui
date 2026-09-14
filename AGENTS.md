# tasks-tui — agent guide

Go TUI (`tasks-tui`) over the `tasks` tracker, on Bubble Tea v2. It shells out to the
`tasks` binary and reads its JSON; it never opens `tasks/*.md`. Design:
`docs/specs/2026-09-13-tasks-tui-v1-design.md`.

## Session protocol

- Start with `tasks prime`; pick from `tasks ready`; `tasks start <id>` before changing code.
- `tasks note <id> "<one line>"` when scope or understanding changes.
- `tasks park <id> "<next step>" [--waiting-on user]` when setting work down.
- `tasks done <id> "<what landed>"` in the same commit as the code. `tasks check` before every commit.
- Never edit `tasks/*.md` by hand. Full protocol: the tasks skill.

## Process and workspace

This repo adopts the process policy in the tasks skill. The task's `process` field
decides whether brainstorming runs: `direct` executes the scoped task or its reviewed
plan; `planned` requires a reviewed design spec and a reviewed implementation plan
first. Before implementation, state the chosen process and workspace; when process is
unassessed, inspect the task and record `tasks edit <id> --process direct|planned`,
and note why.

Work in a worktree under .worktrees/ (`git worktree add`), then `just setup`. Paths
shown to the user are relative to the main checkout (.worktrees/<name>/…).

## Gates

    just gate

`just check` (seconds): `tools/ops-check`, `gofmt`, `go vet`, `staticcheck`, `tasks check`.
`just test`: `go test ./...` — the integration test skips when `tasks` is not on `PATH`.
Hooks: `git config core.hooksPath .githooks` (done by `just setup`).
Rebuild after changes: `just install`.

## Layout

- `cmd/tasks-tui/` — flags, config, startup probe, program.
- internal/tasksctl/ — the only subprocess boundary: `Runner`, typed calls, JSON types, `Error`, `CheckoutFor`.
- internal/identity/ — familiar's slot algorithm (remote key, pins, fnv1a32, hue table, tone).
- internal/quickadd/ — the one-line grammar.
- internal/launch/ — harness commands and the terminal spawn.
- internal/config/ — `~/.config/tasks-tui/config.toml`.
- internal/ui/ — the Bubble Tea model: app, views, overlays, styles.

## Rules

- The JSON contract belongs to `tasks`; anything the TUI needs and cannot get is a task there, never a workaround here.
- A missing named field is a decode error. Fail early with a typed error; no silent fallbacks.
- No `--agent` on `add`: a person files from the TUI.
- Conventional commits; no AI-attribution trailers.
