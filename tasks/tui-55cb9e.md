---
id: tui-55cb9e
title: "Note, block/unblock and shelve/unshelve from the TUI"
status: todo
priority: 2
size: m
complexity: mid
process: direct
created: 2026-10-17T20:36:26Z
updated: 2026-10-17T20:36:26Z
depends: []
parent: tui-d72add
tags: []
agent: claude-code
---

Why: a person records a note, blocks a task, or shelves an idea without leaving the TUI.

Approach: three overlays in the §5.4 pattern (docs/specs/2026-09-13-tasks-tui-v1-design.md), each running one command in the task's checkout (§4.1) and reloading on success:
- `n` note — a required one-line text; `note <id> <text>`.
- `b` block — an optional message; `block <id> [message]`. On a blocked task, `b` offers unblock instead (`unblock <id>`).
- `z` shelve — a required wake condition; `shelve <id> <wake>`. On a shelved task, `z` offers unshelve (`unshelve <id>`). Shelved tasks are not listed in any tab, so unshelve acts from the Task view (`tasks-tui <id>`); a Shelved tab is out of scope.
An error shows `Error.Detail` and keeps the overlay open, as for park.

Done: the overlays work from the Project rows and the Task view; tests in internal/ui/transitions_test.go and internal/tasksctl/writes_test.go cover each command line and a refusal; the three keys are rows in ops tools/keys.toml (landed in ops, re-vendored here) and the keys test passes; the legend lists them; `just gate` passes.

Where: internal/ui/transitions.go, overlay.go, keys.go, legend.go; internal/tasksctl/client.go; tools/keys.toml (owned by ops).
