---
id: tui-cffb85
title: Decide table tunables and mouse input after v1.1
status: done
priority: 2
created: 2026-09-19T15:12:18Z
updated: 2026-09-19T15:25:34Z
completed: 2026-09-19T15:25:34Z
depends: []
tags: []
source: docs/notes/2026-09-19-table-config-and-mouse-brief.md
agent: claude-code/claude-opus-5
---

v1.1 polish (docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md §2, §12) decided the column widths, drop order, and the Projects-pane ready cap in the spec, added no config keys, and left mouse input out; it filed the deferrals as ideas. This goal holds the decisions those ideas wait on: whether the table gains config keys at all (and under one namespace), and whether a keyboard-driven TUI takes pointer input for selection and scrolling. Brief: docs/notes/2026-09-19-table-config-and-mouse-brief.md. Closes when each child idea is scoped, shelved, or dropped.

## Notes

- 2026-09-19T15:25:34Z (main): done
  provenance: {"harness_session":"claude-code:08d0b625-1e1b-4468-a45e-99cd115bfedf","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-09-19T15:25:34Z (main): Decided 2026-09-19: no display config and keyboard only for now; tui-6159db, tui-729559, tui-811ec5 shelved with wake conditions, tui-c2e022 dropped; recorded in docs/notes/2026-09-19-table-config-and-mouse-brief.md
  provenance: {"harness_session":"claude-code:08d0b625-1e1b-4468-a45e-99cd115bfedf","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
