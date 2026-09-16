---
id: tui-e0aa09
title: Remove cursor-timer waits from UI tests
status: doing
priority: 2
size: s
complexity: low
process: direct
owner: main
created: 2026-09-14T13:07:32Z
updated: 2026-09-16T00:04:16Z
started: 2026-09-16T00:04:16Z
depends: []
tags: [v2]
agent: codex
---

Why: UI tests synchronously execute Bubble textinput virtual-cursor blink commands, adding 530ms per typed key. Done: keep blinking enabled by default in production and disable it through the existing UI styles setup used by testEnv; retain synchronous execution of all load and write commands. Where: internal/ui/styles.go, overlay.go, fake_test.go, and focused UI tests. Check: prove prompt input emits no delayed command in the test configuration while production styles retain blink; compare fresh UI test timing before and after.

## Notes

- 2026-09-16T00:04:01Z (main): scope: scoped; direct mechanical test-runtime fix using the UI styles setup, preserving production cursor blinking and synchronous driver coverage
