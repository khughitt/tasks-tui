---
id: tui-c82a86
title: Retain a repeatable terminal smoke test for the complete app
status: todo
priority: 2
size: m
complexity: mid
process: direct
created: 2026-09-14T13:07:32Z
updated: 2026-09-16T01:35:10Z
depends: []
tags: [v2]
agent: codex
---

Why: preserve a real PTY acceptance check beyond the synchronous model integration test. Done: a standalone tools smoke command builds a temporary binary, isolates registry/state and inherited task variables, drives startup/navigation, a write, quick-add paste/cancel, and a fixture launch through tmux, then asserts only scratch paths changed. Where: tools/ smoke command and justfile documentation/recipe if useful. Check: run the command from a clean checkout; it must leave no project registry or state outside its temporary directories. Direct process: the existing real-binary integration setup and prior tmux smoke establish the bounded approach.

## Notes

- 2026-09-16T01:35:10Z (main): scope: scoped; direct standalone tmux smoke reusing the integration test isolation pattern, with a temporary launch fixture and no fast-suite integration
