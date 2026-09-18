---
id: tui-de1545
title: Add a shared task form for editing and creation
status: done
priority: 2
size: m
complexity: mid
process: direct
owner: main
created: 2026-09-17T17:28:32Z
updated: 2026-09-18T21:13:44Z
started: 2026-09-18T21:09:49Z
completed: 2026-09-18T21:13:44Z
depends: []
tags: [v2]
model: "claude-opus-5[1m]"
agent: codex
---

Why:
Tasks can be viewed and transitioned in the TUI, but their core content cannot be edited there. Creation still uses a dense one-line grammar instead of a discoverable form.

Done:
- Pressing `e` for the focused task opens a centered edit form after loading the full task.
- The form always shows exactly three fields in this order: title, body, tags. Title is focused on open and cannot be blank; body and tags may be empty.
- Tab and shift-tab move focus between fields, ctrl+s saves, and esc cancels. Enter remains available for newlines in the body.
- Edit mode prefills all three fields. Saving runs one `tasks edit` call that replaces title, body, and the complete tag list; an error keeps the form open and visible, while success closes it and reloads the current view.
- Pressing `a` opens the same centered form in create mode with blank fields and title focused. Saving files into the current view project with `tasks add`.
- The tags field keeps project tag suggestions. The one-line quick-add overlay, grammar, preview, and any now-unused code are removed once create mode replaces them.
- Priority, size, complexity, project selection, and an advanced-fields toggle are outside this version.

Where to look:
- `internal/ui/app.go`, `internal/ui/keys.go`, `internal/ui/overlay.go`, and `internal/ui/quickadd.go` contain global bindings and the current bottom overlay.
- `internal/tasksctl/client.go` is the only subprocess boundary and needs the typed edit call.
- Bubble Tea Bubbles already provides `textinput` and `textarea`; add no form dependency.
- `tui-d72add` is the broader v2 idea for other field edits and task actions.

Correctness:
UI tests cover centered rendering, initial title focus, field order and focus movement, edit prefill and argv, create argv, tag suggestions, validation, and keeping failed submissions open. Run `just gate` and rebuild with `just install`.

## Notes

- 2026-09-17T17:29:57Z (main): scope: scoped; promoted to a direct medium task with edit mode first and create mode replacing quick add second
- 2026-09-18T21:09:49Z (main): started
  provenance: {"harness_session":"claude-code:d7132380-8d35-431d-99b2-4362c33f819e","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-18T21:13:44Z (feat/task-form): done
  provenance: {"harness_session":"claude-code:d7132380-8d35-431d-99b2-4362c33f819e","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-18T21:13:44Z (feat/task-form): Shared centered title/body/tags form edits tasks and replaces quick add for creation, with tag completion and typed add/edit writes; load warnings are the form's own, the tag lookup's are dropped
  provenance: {"harness_session":"claude-code:d7132380-8d35-431d-99b2-4362c33f819e","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
