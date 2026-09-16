---
id: tui-62bb3f
title: Sort task lists by column
status: done
priority: 2
size: m
complexity: mid
process: planned
owner: feat/task-list-sorting
created: 2026-09-15T02:17:13Z
updated: 2026-09-16T10:16:11Z
started: 2026-09-16T02:34:12Z
completed: 2026-09-16T10:16:11Z
depends: []
tags: [v2]
agent: codex
spec: docs/specs/2026-09-15-task-list-sorting-design.md
plan: docs/plans/2026-09-16-task-list-sorting.md
step: "Task 1: Local sort model and measured headers"
---

Let people reorder the current Project tab by its visible task-table columns without issuing a new tasks command. Preserve filtering, tabs, selection, and the existing server-defined default order on initial load. Decide and document the keyboard interaction, sort direction, sortable fields, indicators, and tests before implementation.

## Notes

- 2026-09-16T02:32:15Z (main): scope: scoped; planned because the visible-column sort interaction, direction, and supported fields are not settled by the v1.1 spec
- 2026-09-16T02:32:54Z (feat/task-list-sorting): parked (waiting on user, decision): Choose the task-list sort interaction: recommended keyboard header selector versus a fixed key cycle
- 2026-09-16T02:34:12Z (feat/task-list-sorting): design decision: approved keyboard header selector; arrows choose a visible data column and Enter toggles direction
- 2026-09-16T02:59:37Z (feat/task-list-sorting): design revised after review: ranked enums, selector capture and off cycle, measured indicators, and reload/tab behavior are explicit
- 2026-09-16T09:02:31Z (feat/task-list-sorting): design approved: hidden-column behavior clarified; proceed to implementation planning
- 2026-09-16T09:45:38Z (feat/task-list-sorting): plan refined: each child completion stages its task record; feedback tasks-136399 records the tracker CLI link-clearing gap
- 2026-09-16T10:16:11Z (feat/task-list-sorting): Added local keyboard sorting for Project task lists
