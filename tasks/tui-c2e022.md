---
id: tui-c2e022
title: Map a Bubble Tea v2 click or wheel event to a task row and column
status: dropped
priority: 2
size: s
complexity: mid
process: direct
created: 2026-09-19T15:12:45Z
updated: 2026-09-19T15:25:34Z
depends: []
parent: tui-cffb85
tags: [v2]
agent: claude-code/claude-opus-5
---

Question: Can the TUI, with View.MouseMode set to MouseModeCellMotion, turn a tea.MouseClickMsg into the selected row (and the clicked column key) and a tea.MouseWheelMsg into a scroll, using only what the column model already measures?

Where to start: charm.land/bubbletea/v2 mouse.go (MouseClickMsg, MouseWheelMsg, Mouse.X/Y) and tea.go (View.MouseMode); internal/ui/app.go View() and the key dispatch at key(); internal/ui/columns.go offset() and internal/ui/rows.go titleOffset() for column geometry; internal/ui/project.go render() for the row offset/sel arithmetic; internal/ui/projects.go for the two-table layout on the Projects view. Linked from docs/notes/2026-09-19-table-config-and-mouse-brief.md.

Bound: A throwaway spike in a worktree that enables the mouse mode and logs the row and column under one click and the direction of one wheel event in the project view and the Projects view; no selection, scrolling, or sort behaviour is implemented and nothing is committed to main.

Expected result: Record on this task and in the brief whether row and column resolve from (X, Y) with the existing offsets or need the views to keep the header height and per-row y they render, whether the alt screen and the overlays interfere, and a recommendation on whether mouse selection is a bounded direct task.

Ideas it wakes: On completion, run tasks note on tui-811ec5 with the finding, in the same commit as this result.

## Notes

- 2026-09-19T15:25:34Z (main): dropped
  provenance: {"harness_session":"claude-code:08d0b625-1e1b-4468-a45e-99cd115bfedf","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-19T15:25:34Z (main): Keyboard-only decided 2026-09-19 on the brief; the hit-testing question has no consumer until tui-811ec5 is unshelved
  provenance: {"harness_session":"claude-code:08d0b625-1e1b-4468-a45e-99cd115bfedf","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
