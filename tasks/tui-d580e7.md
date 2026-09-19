---
id: tui-d580e7
title: Animate list transitions with harmonica
status: shelved
priority: 2
created: 2026-09-15T02:17:13Z
updated: 2026-09-19T15:12:32Z
depends: []
tags: [v2]
agent: codex
---

Deferred from v1.1 polish (§12); lists currently move only for the loading spinner.

## Notes

- 2026-09-19T15:12:32Z (main): shelved: A named transition (tab switch, pane reflow, or sort) is observed to read as a jump; v1.1 §12 holds that nothing in a list of rows earns motion beyond the spinner
- 2026-09-19T15:12:32Z (main): scope: shelved; the v1.1 spec (§12) rejects motion in row lists on principle and no transition has been named as jarring; related to tui-cffb85 through the brief but left unparented so the shelf does not hold the goal open; brief: docs/notes/2026-09-19-table-config-and-mouse-brief.md
