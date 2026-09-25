---
id: tui-656abf
title: Replace hard-coded colors with theme-following ANSI roles
status: todo
priority: 2
size: m
complexity: mid
process: planned
created: 2026-09-25T12:27:49Z
updated: 2026-09-25T12:27:49Z
depends: []
tags: []
source: tasks-142d2f follow-up
agent: claude-code/claude-opus-5-5
---

internal/ui/styles.go:40-52 fixes its palette as xterm-256 indices (red 124/203, orange 166/215, yellow 136/221, teal 30/80, dim 245/243) and truecolor pill text (#fafafa/#101010), so the TUI ignores the terminal theme. Replace them with basic ANSI slots (30-37, dim, bold), or derive them from the terminal's own palette the way tasks now does for dates (tasks docs/specs/2026-09-25-date-recency-color-design.md: OSC 10/11/4 query, TASKS_PALETTE override). Out of scope: the per-project identity tones in internal/identity/tone.go, which are hashed HSL hues by design; decide explicitly whether they stay.
