---
id: tui-d6e352
title: "tasks-tui v1: the daily surface"
status: doing
priority: 1
size: xl
complexity: high
process: planned
owner: main
created: 2026-09-14T00:45:49Z
updated: 2026-09-14T00:45:49Z
started: 2026-09-14T00:45:49Z
depends: []
tags: [v1]
agent: claude-code/claude-opus-5
---

A standalone Bubble Tea TUI over the tasks tracker: multi-project and single-project views, task detail, quick add with inline token syntax, quick launch of an agent session on a task, and the core transitions start/park/done/drop. Shells out to the tasks binary and reads JSON; never reads or writes tasks/*.md. Per-project accent follows familiar's slot algorithm. Subsumes tasks-202e1f (quick launch).
