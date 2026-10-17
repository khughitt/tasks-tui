---
id: tui-c30d88
title: Design the tree and dependency views from the brief
status: todo
priority: 2
size: m
complexity: high
process: direct
created: 2026-10-17T20:35:34Z
updated: 2026-10-17T20:35:34Z
depends: []
parent: tui-29c34d
tags: []
source: docs/notes/2026-10-17-tasks-tui-v2-surface-brief.md
agent: claude-code
---

Question: Which tree and dependency views should tasks-tui have — a Tree tab from `tasks tree --project`, a per-task neighbourhood (depends on / depended on by) from list-row `depends`, a whole-project graph, or a combination — and where do they sit among the views and keys?
Where to start: the brief's §4 alternatives (docs/notes/2026-10-17-tasks-tui-v2-surface-brief.md); `tasks tree --project tui` JSON; list-row `depends`; v1 spec §5 and v1.1 spec §3 (column model); tasks-e8d471 (dependents in show).
Bound: A draft design spec under docs/specs/ with Status: draft, reviewed with the user; no implementation. Keys proposed for ops tools/keys.toml, not landed.
Expected result: The reviewed spec attached to tui-bbd26e with --spec, the chosen alternative recorded on this task and in the brief.
Ideas it wakes: On completion, run `tasks note tui-bbd26e --no-reviewed "<chosen views and spec path>"` before this result's branch merges.
