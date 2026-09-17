---
id: tui-bbcb55
title: Accept sparse JSON task records from tasks CLI
status: done
priority: 2
size: s
complexity: low
process: direct
owner: fix/sparse-json
created: 2026-09-17T16:37:29Z
updated: 2026-09-17T16:40:10Z
started: 2026-09-17T16:37:40Z
completed: 2026-09-17T16:40:10Z
depends: []
tags: []
agent: codex
---

Companion to tasks-82d845, explicitly authorized by user. Keep required scalar/envelope validation, allow omitted optional task fields and empty task collections under the new tasks contract, cover decoder and integration paths, update contract descriptions and reinstall TUI.

## Notes

- 2026-09-17T16:38:25Z (fix/sparse-json): took over session sid:1099759 (owner main, host titan, pid 1099759, worktree /mnt/ssd/Dropbox/tasks-tui, since 2026-09-17T16:37:40Z, age 45s, stale: pid 1099759 is gone)
- 2026-09-17T16:38:25Z (fix/sparse-json): Direct process in .worktrees/sparse-json: user authorized companion decoder update; sparse fields default naturally through Go JSON pointers/slices, required scalars and envelopes stay validated.
- 2026-09-17T16:40:10Z (fix/sparse-json): Verified just gate with PATH pointing to sparse tasks debug binary, including real CLI integration; independent code review found no issues.
- 2026-09-17T16:40:10Z (fix/sparse-json): Accept omitted optional task fields and collections while preserving required-key validation; regression and real-CLI integration pass.
