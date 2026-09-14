---
id: tui-c82a86
title: Retain a repeatable terminal smoke test for the complete app
status: idea
priority: 2
created: 2026-09-14T13:07:32Z
updated: 2026-09-14T13:07:32Z
depends: []
tags: [v2]
agent: codex
---

The v1 acceptance pass exercised the installed binary in a real PTY, but its controller scratch script was removed after review. The committed integration test drives the model synchronously and substitutes Spawn, so it does not retain terminal startup, escape sequences, or actual detached launch coverage. Add a small repeatable smoke command covering startup/navigation, one write, prompt paste/cancellation, and a harmless launch fixture. Isolate XDG_CONFIG_HOME and XDG_STATE_HOME, scrub inherited TASKS_* variables, and verify no real registry/state pollution. Prefer a small standard-library script and keep it separate from the fast suite.
