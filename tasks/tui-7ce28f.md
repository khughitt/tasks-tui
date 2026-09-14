---
id: tui-7ce28f
title: "v1.1: UI polish — adaptive columns, marks column, tinted selection, help panel"
status: doing
priority: 2
size: l
complexity: mid
process: planned
owner: feat/ui-polish
created: 2026-09-14T23:20:08Z
updated: 2026-09-14T23:20:18Z
started: 2026-09-14T23:20:18Z
depends: []
tags: [ui]
agent: "claude-code/claude-opus-5[1m]"
spec: docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md
---

Polish pass over every view after v1 landed. Alignment: the Projects table pads with literal widths and the counts are left-aligned strings; every task row column is a literal pad. Feel: full-reverse selection, a hard-coded help string, no motion while loading. Keys: q is swallowed by the picker and confirm overlays; enter is the only open key. Spec: docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md.
