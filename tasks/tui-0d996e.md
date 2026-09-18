---
id: tui-0d996e
title: Hide/remove the parked task / reusme messages that appear in the bottom of the TUI.
status: done
priority: 2
process: direct
owner: main
created: 2026-09-17T17:17:56Z
updated: 2026-09-18T20:59:30Z
started: 2026-09-18T20:36:13Z
completed: 2026-09-18T20:59:30Z
depends: []
tags: []
model: "claude-opus-5[1m]"
---

## Notes

- 2026-09-18T20:36:13Z (main): started
- 2026-09-18T20:36:22Z (main): process direct: single scoped UI tweak, settle via code reading
- 2026-09-18T20:48:51Z (main): resumed
  provenance: {"harness_session":"claude-code:73fd2c65-9871-48c7-ae8d-afafae6a6a85","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-18T20:48:52Z (main): scoping: the clutter is load warnings (prime/list/show) re-entering the status line on every 30s reload; design a quiet badge + console instead of suppressing them
- 2026-09-18T20:59:30Z (feat/warning-console): done
  provenance: {"harness_session":"claude-code:73fd2c65-9871-48c7-ae8d-afafae6a6a85","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-18T20:59:30Z (feat/warning-console): load warnings are view data behind a ⚠ N status-line badge; W opens the console (current warnings, then the log); startup and tag-lookup warnings dropped; spec §11 amended
  provenance: {"harness_session":"claude-code:73fd2c65-9871-48c7-ae8d-afafae6a6a85","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
