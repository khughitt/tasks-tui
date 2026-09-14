---
id: tui-e0aa09
title: Remove cursor-timer waits from UI tests
status: idea
priority: 2
created: 2026-09-14T13:07:32Z
updated: 2026-09-14T13:07:32Z
depends: []
tags: [v2]
agent: codex
---

The v1 UI suite took about 49 seconds; the synchronous uitest driver executes tea.Cmd directly, while the installed cursor implementation waits up to 530ms per blink command. Investigate and remove animation waits from tests using the smallest existing cursor/test configuration mechanism. Preserve per-key and paste assertions, load/write command execution, and production cursor behavior. Compare fresh UI/integration timings before and after; avoid a general-purpose scheduler unless necessary.
