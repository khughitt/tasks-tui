---
id: tui-c82a86
title: Retain a repeatable terminal smoke test for the complete app
status: done
priority: 2
size: m
complexity: mid
process: direct
owner: test/terminal-smoke
created: 2026-09-14T13:07:32Z
updated: 2026-09-16T01:42:49Z
started: 2026-09-16T01:35:28Z
completed: 2026-09-16T01:42:49Z
depends: []
tags: [v2]
agent: codex
---

Why: preserve a real PTY acceptance check beyond the synchronous model integration test. Done: a standalone tools smoke command builds a temporary binary, isolates registry/state and inherited task variables, drives startup/navigation, a write, quick-add paste/cancel, and a fixture launch through tmux, then asserts only scratch paths changed. Where: tools/ smoke command and justfile documentation/recipe if useful. Check: run the command from a clean checkout; it must leave no project registry or state outside its temporary directories. Direct process: the existing real-binary integration setup and prior tmux smoke establish the bounded approach.

## Notes

- 2026-09-16T01:35:10Z (main): scope: scoped; direct standalone tmux smoke reusing the integration test isolation pattern, with a temporary launch fixture and no fast-suite integration
- 2026-09-16T01:41:07Z (test/terminal-smoke): Standalone tmux smoke now builds a temporary binary in an env -i scratch project, verifies a pasted quick-add write, raw-escape cancellation, and the detached fake harness without HOME fallback writes
- 2026-09-16T01:42:49Z (test/terminal-smoke): Add a repeatable tmux smoke command covering real PTY startup/navigation, pasted quick-add write/cancel, detached fixture launch, and isolated XDG state
