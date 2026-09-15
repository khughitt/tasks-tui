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
updated: 2026-09-15T01:07:59Z
started: 2026-09-14T23:20:18Z
depends: []
tags: [ui]
agent: "claude-code/claude-opus-5[1m]"
spec: docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md
plan: docs/plans/2026-09-14-tasks-tui-v1.1-polish.md
---

Polish pass over every view after v1 landed. Alignment: the Projects table pads with literal widths and the counts are left-aligned strings; every task row column is a literal pad. Feel: full-reverse selection, a hard-coded help string, no motion while loading. Keys: q is swallowed by the picker and confirm overlays; enter is the only open key. Spec: docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md.

## Notes

- 2026-09-14T23:20:26Z (feat/ui-polish): parked (waiting on user, review): User reviews the v1.1 polish spec; on approval write the implementation plan (writing-plans) with one step child per section
- 2026-09-14T23:25:06Z (feat/ui-polish): Design review: define overflow after all optional columns drop and reconcile pane/table minimum widths; make stacked help reachable within terminal height; connect flexCell and mixed-style marks to the row API; separate header labels from structural column keys.
- 2026-09-14T23:27:23Z (feat/ui-polish): Review round 1: four findings (narrow floors, legend height, cell API, header labels) resolved in the spec; awaiting re-review
- 2026-09-15T00:11:22Z (feat/ui-polish): parked (waiting on user, review): User reviews docs/plans/2026-09-14-tasks-tui-v1.1-polish.md; on approval execute Task 1 (tui-cf870c) onward via subagent-driven development in this worktree
- 2026-09-15T00:16:56Z (feat/ui-polish): Implementation-plan review: extracted column tests pass but row alignment test falsely fails on UTF-8 byte offsets (Projects test has same issue); unreachable indicator is lost when name drops; loaded flag loses prior tab counts; legend fit checks height without width; pill/empty-state test contradicts persistent filters.
- 2026-09-15T01:03:22Z (feat/ui-polish): Plan review round 1: five findings (byte-offset asserts, unreachable indicator, tab counts, legend width, filter persistence) resolved; awaiting re-review
- 2026-09-15T01:06:52Z (feat/ui-polish): Re-review of 976e5ca: prior five findings addressed; new Task 5 compile blocker at plan lines 1288-1289: tabNames is []string, so [len(tabNames)] array lengths are not constant. Compiler reproduction fails; changing tabNames to [...]string compiles. Add that declaration change to the plan.
- 2026-09-15T01:07:59Z (feat/ui-polish): Resolved the final plan-review finding: Task 5 explicitly changes tabNames to an inferred-length array before declaring shown/hasCount; corrected its file summary. Implementation remains pending.
