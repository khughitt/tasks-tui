---
id: tui-62bb3f
title: Sort task lists by column
status: doing
priority: 2
size: m
complexity: mid
process: planned
owner: feat/task-list-sorting
created: 2026-09-15T02:17:13Z
updated: 2026-09-16T02:34:12Z
started: 2026-09-16T02:34:12Z
depends: []
tags: [v2]
agent: codex
---

Let people reorder the current Project tab by its visible task-table columns without issuing a new tasks command. Preserve filtering, tabs, selection, and the existing server-defined default order on initial load. Decide and document the keyboard interaction, sort direction, sortable fields, indicators, and tests before implementation.

## Notes

- 2026-09-16T02:32:15Z (main): scope: scoped; planned because the visible-column sort interaction, direction, and supported fields are not settled by the v1.1 spec
- 2026-09-16T02:32:54Z (feat/task-list-sorting): parked (waiting on user, decision): Choose the task-list sort interaction: recommended keyboard header selector versus a fixed key cycle
- 2026-09-16T02:34:12Z (feat/task-list-sorting): design decision: approved keyboard header selector; arrows choose a visible data column and Enter toggles direction
