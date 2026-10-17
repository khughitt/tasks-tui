---
id: tui-d72add
title: "Field editing, dep, shelve/unshelve, note, block from the TUI"
status: todo
priority: 2
size: l
complexity: mid
process: direct
created: 2026-09-14T12:42:09Z
updated: 2026-10-17T20:36:05Z
depends: []
parent: tui-29c34d
tags: [v2]
agent: codex
---

Deferred from v1 (docs/specs/2026-09-13-tasks-tui-v1-design.md §14): editing the fields beyond title, body, and tags; dependencies (`dep`), shelve/unshelve, `note`, `block`.

Why: these are the remaining writes a person makes on a task; today each needs the CLI.

Established 2026-10-17: every write exists in the tasks CLI — `edit -p/--size/--complexity/--process/--every/--defer/--parent/--need`, `dep <id> --on/--rm`, `note <id> <text>`, `block <id> [message]`, `unblock`, `shelve <id> <wake>`, `unshelve`. No tasks contract is missing. The TUI wires only `edit --title -b --no-tags --tag` (internal/tasksctl/client.go). Every new key is a row in ops tools/keys.toml, landed in ops and re-vendored before the keys test passes here.

Done: the three children are closed: one-line overlays (note, block/unblock, shelve/unshelve), rubric fields in the task form, and dependency editing.

Where: internal/ui/transitions.go and overlay.go (the §5.4 overlay pattern), internal/ui/task_form.go (§6), internal/tasksctl/client.go, tools/keys.toml; brief docs/notes/2026-10-17-tasks-tui-v2-surface-brief.md.
