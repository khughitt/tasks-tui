---
id: tui-ce6c9f
title: Rename the Go module to github.com/khughitt/tasks-tui so go install @latest works
status: done
priority: 2
size: s
complexity: low
process: direct
owner: main
created: 2026-09-19T13:13:54Z
updated: 2026-09-19T14:29:34Z
started: 2026-09-19T14:28:25Z
completed: 2026-09-19T14:29:34Z
depends: []
tags: []
model: "claude-opus-5[1m]"
agent: "claude-code/claude-opus-5[1m]"
---

go.mod declares module tasks-tui, so the public repo cannot be installed with go install github.com/khughitt/tasks-tui/cmd/tasks-tui@latest; only a checkout plus just install works. Rename the module path, rewrite the internal imports, keep the tui symlink recipe, and add the go install line to the README Install section.

## Notes

- 2026-09-19T14:28:25Z (main): started
  provenance: {"harness_session":"claude-code:5a134633-4e2c-4d08-9c66-62543b270238","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-19T14:29:34Z (feat/module-path): done
  provenance: {"harness_session":"claude-code:5a134633-4e2c-4d08-9c66-62543b270238","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-19T14:29:34Z (feat/module-path): Module is github.com/khughitt/tasks-tui; imports rewritten; README Install leads with go install @latest, just install keeps the tui symlink
  provenance: {"harness_session":"claude-code:5a134633-4e2c-4d08-9c66-62543b270238","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
