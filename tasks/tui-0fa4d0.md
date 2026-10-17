---
id: tui-0fa4d0
title: Read ops identity.toml for project names and purposes
status: idea
priority: 2
created: 2026-09-14T12:42:09Z
updated: 2026-10-17T21:04:21Z
depends: []
tags: [v2]
agent: codex
---

Deferred from v1 (docs/specs/2026-09-13-tasks-tui-v1-design.md §14). Show canonical project names and purposes in the Projects view instead of the root basename. Handoff: docs/notes/2026-09-15-project-identity-brief.md.

## Open questions

- Should `tasks projects` expose `name` and `purpose` (read from each root's `identity.toml`, or from ops's `identity-mirror.toml`), so the TUI only decodes them? Recommendation: yes, from the root's `identity.toml`, optional until all projects adopt one (16 of 34 on 2026-10-17). If yes, file the field task in `tasks` and a dependent TUI consumer task here.
- Where should a purpose render: the selected-project pane header (lean), a wide-layout column, or task detail?

## Notes

- 2026-09-15T21:17:26Z (main): scope: briefed; current table only derives a basename and the canonical identity file has no settled TUI contract or purpose placement; brief: docs/notes/2026-09-15-project-identity-brief.md
- 2026-10-17T20:58:36Z (main): scope: question; brief updated for per-project identity.toml + identity-mirror.toml (ops-03f3e3), 16/34 roots adopted, tasks 0.4.0 projects JSON still has no name/purpose, no tasks-project task for it; questions moved to body; brief: docs/notes/2026-09-15-project-identity-brief.md
- 2026-10-17T21:04:21Z (main): answer (user, 2026-10-17): 1. yes — tasks projects exposes name and purpose, from each root's identity.toml, optional until adoption is complete; 2. the purpose renders at the head of the selected project's pane
