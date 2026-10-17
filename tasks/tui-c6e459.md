---
id: tui-c6e459
title: Watch tasks/ with inotify instead of the refresh timer
status: idea
priority: 2
reviewed: true
awaiting_user: true
reviewed_at: 2026-10-17T20:37:04Z
created: 2026-09-14T12:42:09Z
updated: 2026-10-17T20:37:04Z
depends: []
parent: tui-29c34d
tags: [v2]
agent: codex
---

Deferred from v1 (docs/specs/2026-09-13-tasks-tui-v1-design.md §14).

Evidence 2026-10-17: refresh is a `tea.Tick` every `refresh_seconds` (default 30; internal/ui/app.go). Claims and parks, the main reason to refresh (v1 §10), live in tasks' out-of-git claim store, not under `tasks/`, and the TUI must not read tasks' state paths, so watching `tasks/` would miss them; the Projects view would also have to watch every registered root and worktree. Cost per tick: `prime --project` 0.01 s, `tasks projects` 0.19 s, `list --all-projects --status doing` 1.82 s (the Projects strip runs it and `--parked`). Alternatives and lean: docs/notes/2026-10-17-tasks-tui-v2-surface-brief.md §4 — keep the timer (slower strip), or a tasks-side change feed filed in tasks; a TUI-side watch is rejected.

## Open questions

- Is the up-to-30 s lag behind agents, or the Projects view's ~2 s poll every tick, felt in use? If neither, shelve this; if the lag is, file a change feed (`tasks watch` or a per-project change stamp) in tasks and have this wait on it.

## Notes

- 2026-10-17T20:37:04Z (main): scope: question; inotify on tasks/ misses claim/park changes held in the claim store, so the idea as titled is rejected; whether any push refresh is wanted depends on whether the lag is felt; parented under tui-29c34d; brief: docs/notes/2026-10-17-tasks-tui-v2-surface-brief.md
