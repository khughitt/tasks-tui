---
id: tui-0fa4d0
title: Read ops identity.toml for project names and purposes
status: todo
priority: 2
size: s
complexity: mid
process: direct
created: 2026-09-14T12:42:09Z
updated: 2026-10-17T21:05:32Z
depends: [tasks-6845e6]
tags: [v2]
agent: codex
---

Deferred from v1 (docs/specs/2026-09-13-tasks-tui-v1-design.md §14). Handoff: docs/notes/2026-09-15-project-identity-brief.md.

Why: the Projects view shows each root's basename (`name()` in internal/ui/projects.go) where projects have canonical names and one-line purposes. Decided by the user 2026-10-17: `tasks projects` carries `name` and `purpose` from each root's `identity.toml` (tasks-6845e6, which this depends on), optional until every project adopts one; the purpose renders at the head of the selected project's pane.

Done:
- `tasksctl` decodes the optional `name` and `purpose` on a projects row (absent is nil; types.go beside `Prefix`/`Root`); fixture `projects.json` recaptured with `just capture projects projects`.
- The `name` column and the filter use `name` when present and the root basename when absent, as an explicit rule tested both ways.
- `renderPane` gains a first line with the purpose (muted, cut to the pane width), shown also when the pane is otherwise empty (the early "nothing doing, parked, or ready" return), and omitted when the project has no purpose.
- Narrow layouts keep dropping the `name` column as now; the pane line survives because it is not a column.
- `just gate` green.

Where: internal/ui/projects.go (`name`, filter at the rows build, `renderPane`, the column table), internal/tasksctl/types.go and testdata/projects.json, README if it describes the name column.

## Notes

- 2026-09-15T21:17:26Z (main): scope: briefed; current table only derives a basename and the canonical identity file has no settled TUI contract or purpose placement; brief: docs/notes/2026-09-15-project-identity-brief.md
- 2026-10-17T20:58:36Z (main): scope: question; brief updated for per-project identity.toml + identity-mirror.toml (ops-03f3e3), 16/34 roots adopted, tasks 0.4.0 projects JSON still has no name/purpose, no tasks-project task for it; questions moved to body; brief: docs/notes/2026-09-15-project-identity-brief.md
- 2026-10-17T21:04:21Z (main): answer (user, 2026-10-17): 1. yes — tasks projects exposes name and purpose, from each root's identity.toml, optional until adoption is complete; 2. the purpose renders at the head of the selected project's pane
