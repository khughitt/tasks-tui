---
id: tui-c2a9df
title: Add and remove dependencies from the TUI
status: todo
priority: 2
size: m
complexity: mid
process: planned
created: 2026-10-17T20:36:29Z
updated: 2026-10-17T20:36:29Z
depends: []
parent: tui-d72add
tags: []
agent: claude-code
---

Why: a person records that one task waits on another while looking at both, without composing `tasks dep`.

Context: `dep <id> --on <ref>... --rm <ref>...` applies both in one save; refs may cross projects. The Task view lists dependencies from `show` but not dependents (tasks-e8d471 is open in tasks). tasks-9c18e7 (dep accepts an edge to the task's own ancestor) may change what `dep` refuses. The design decides how a target is chosen (typed id with completion, or a picker over list rows), whether removal acts on the relations block, and whether cross-project refs are offered. It should follow the tree/dependency view design (tui-c30d88) where the two overlap.

Done: a reviewed design and plan; adding and removing a dependency from the Task view runs one `dep` command in the task's checkout and reloads; a refusal is shown; tests cover the command lines; `just gate` passes.

Where: internal/ui/task.go (relations block), transitions.go, internal/tasksctl/client.go.
