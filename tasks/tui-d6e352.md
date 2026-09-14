---
id: tui-d6e352
title: "tasks-tui v1: the daily surface"
status: done
priority: 1
size: xl
complexity: high
process: planned
owner: design/v1
created: 2026-09-14T00:45:49Z
updated: 2026-09-14T12:52:58Z
started: 2026-09-14T00:45:49Z
completed: 2026-09-14T12:52:58Z
depends: []
tags: [v1]
agent: claude-code/claude-opus-5
spec: docs/specs/2026-09-13-tasks-tui-v1-design.md
plan: docs/plans/2026-09-13-tasks-tui-v1.md
---

A standalone Bubble Tea TUI over the tasks tracker: multi-project and single-project views, task detail, quick add with inline token syntax, quick launch of an agent session on a task, and the core transitions start/park/done/drop. Shells out to the tasks binary and reads JSON; never reads or writes tasks/*.md. Per-project accent follows familiar's slot algorithm. Subsumes tasks-202e1f (quick launch).

## Notes

- 2026-09-14T00:52:12Z (design/v1): parked (waiting on user, review): User reviews docs/specs/2026-09-13-tasks-tui-v1-design.md on branch design/v1 (.worktrees/v1); then writing-plans
- 2026-09-14T01:11:38Z (design/v1): Spec review round 1: six findings verified against tasks src and familiar pins.js; all six revised in
- 2026-09-14T01:15:51Z (design/v1): Spec review round 2: stderr envelope, claim.live, by-id entry via the parked feed; all three revised in
- 2026-09-14T01:51:54Z (design/v1): parked (waiting on user, review): User reviews docs/plans/2026-09-13-tasks-tui-v1.md (19 tasks, chained); then execute Task 1 in .worktrees/v1
- 2026-09-14T03:17:31Z (design/v1): Plan review round 1: seven findings from compiling the plan's code; all revised in (focus routing, nested nullability, notices, merge by id, grammar, fixtures/ANSI, integration env)
- 2026-09-14T09:23:11Z (design/v1): Plan review round 2: unreachable projects (null counts/total), DepInfo nullability, pane scope by prefix, fixture tags; all revised in
- 2026-09-14T10:29:32Z (design/v1): User approved reviewed plan for implementation using subagent-driven-development; executing the 19 direct steps in existing .worktrees/v1.
- 2026-09-14T10:59:01Z (design/v1): Implementation follows the reviewed plan; execution review adds strict error-detail decoding, notices for missing live-claim worktrees, familiar saturation clamping, and small UI corrections for visible selection, UTF-8 filters, header counts, warning sources, and preview colors. Rulings tracked in the SDD ledger.
- 2026-09-14T12:52:58Z (design/v1): Implemented all 19 reviewed plan steps plus the final null-park decoder fix. Independent task reviews and final scoped review are clean; just gate and real-terminal smoke passed, including isolated transitions, quick add, launch, and real registry entry points. Branch design/v1 remains available for review.
- 2026-09-14T12:52:58Z (design/v1): tasks-tui v1: Projects/Project/Task views, quick add, launch, start/park/done/drop over the tasks JSON contract
