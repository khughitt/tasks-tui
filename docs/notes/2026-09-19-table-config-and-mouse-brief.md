# Table config and mouse brief

**Status:** decided 2026-09-19 — no display config and keyboard only for now; both leans below were confirmed. `tui-6159db`, `tui-729559`, `tui-811ec5`, and `tui-d580e7` are shelved with wake conditions, `tui-c2e022` dropped, goal `tui-cffb85` done. The sections below are the case as it was presented.

## Problem

Decide what the v1.1 table surface exposes next: whether the column set, column order,
and the Projects-pane ready cap become config keys, and whether the TUI takes mouse
input for selecting and scrolling rows. v1.1 withheld all three deliberately and filed
them as ideas; nobody has yet asked for any of them.

## Current behaviour and evidence

- `docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md` §2 "No new configuration":
  widths, the pane cap, and the fade thresholds are decided in the spec; "a person
  who wants them tunable files an idea". §12 lists the deferrals. The ideas were filed
  by the plan (`docs/plans/2026-09-14-tasks-tui-v1.1-polish.md`, `tasks add … --status
  idea --tag v2`), not from observed demand.
- The column set is one declaration, `taskTable` in `internal/ui/rows.go`, with a
  drop order (`proc` → `cx` → `sz` → `age`), one flexible column (`title`), and two
  structural columns (`gutter`, `marks`). `internal/ui/columns.go` measures widths over
  every row and exposes `offset(widths, key)`; `rows.go` uses it for `titleOffset()`.
- Sorting (`docs/specs/2026-09-15-task-list-sorting-design.md`) keys chords to column
  keys and renders the direction arrow in the header; it puts mouse header clicks and
  configurable default sorts out of scope.
- `paneReadyCap = 8` in `internal/ui/projects.go` (since `86ba772`), rendered as
  `N ready · 8 shown`.
- Config (`internal/config/config.go`, README "Config") is flat: `tasks`,
  `refresh_seconds`, `[launch]`, `[identity]`. Unknown keys are an error and every key
  is validated.
- Mouse: nothing is enabled. `App.View()` returns `tea.NewView` with `AltScreen`
  only; Bubble Tea v2.0.9 sets mouse capture per view through `View.MouseMode`
  (`MouseModeCellMotion` gives click, release, and wheel) and delivers
  `tea.MouseClickMsg` / `tea.MouseWheelMsg` with cell coordinates. `App.key` dispatches
  only `tea.KeyPressMsg`; chords are prefix-buffered there.
- Animation: no harmonica dependency; the only motion is the loading spinner.

## Constraints

- The v1 design (`docs/specs/2026-09-13-tasks-tui-v1-design.md`) names the TUI
  keyboard-driven; the legend is generated from the keymap and a mouse action has no
  binding to appear there.
- The column model's invariants (widths measured over every row, at most one flexible
  column, a declared drop order, `minWidth` driving the Projects layout) must survive
  any user-chosen column set; a config that names columns must not be able to drop
  `gutter`, `marks`, `id`, or `title`.
- The sort chords, header labels, and hint column in the sorting spec address columns
  by key; a reordered or hidden column must keep those keys stable.
- Config policy: unknown keys fail early, every key validates, defaults live in
  `Default()`; no silent fallbacks.
- Open work: `tui-0fa4d0` (project names from ops) may add a column to the Projects
  table; a columns config would need to know about it.

## Alternatives

**Config (tui-6159db, tui-729559)**

1. Keep v1.1's stance: no display config. Both ideas are dropped when the user confirms
   nobody wants them. Cheapest; consistent with the spec's reasoning.
2. One `[ui]` table: `ui.pane_ready_rows = 8` now, and later `ui.task_columns =
   ["id", "prio", …]` as an ordered subset of the droppable keys. The cap lands as an
   xs direct task; the columns key needs a design that says how a chosen order
   interacts with the drop order and `flexMin`.
3. Per-view tables (`[projects]`, `[project]`). Rejected for now: two namespaces for
   two keys, and the task table is shared by both views.

Lean: 1 unless the user names a use. If a use exists, 2, with the cap first.

**Mouse (tui-811ec5)**

1. None; keyboard only. Consistent with the v1 design and the generated legend.
2. Cell-motion mode with click-to-select and wheel-to-scroll only, no header clicks,
   no drag. Hit-testing needs the row y-range each view renders (header height, the
   Projects table/pane split) plus `offset()` for the column; `tui-c2e022` establishes
   whether the existing geometry suffices.
3. Full pointer UI (header click sorts, tab clicks, overlay buttons). Rejected: it
   duplicates the keymap and the sorting spec kept header clicks out.

Lean: 1 until someone wants to use the TUI from a pointer; 2 is the only step worth
taking if so, and only after `tui-c2e022` reports.

**Animation (tui-d580e7)**: shelved; §12 states the position and no transition has
been named as jarring.

## Unanswered questions

- Does anyone want a different column set, order, or ready cap? The user; the spec's
  "files an idea" was meant as demand, and none has arrived.
- If yes, is a `[ui]` table the namespace? The user, given alternative 2 above.
- Is pointer input wanted in a keyboard-driven TUI at all? The user.
- Can a click resolve to a row and column from what the views already measure, and
  do the alt screen and overlays interfere? `tui-c2e022`.

## Proposed decomposition

- `tui-cffb85` — goal holding the decisions; source is this brief.
- `tui-c2e022` — research: map a v2 click or wheel event to a row and column; wakes
  `tui-811ec5`.
- `tui-729559` waits on the config namespace answer; on a yes it is scoped as xs,
  low complexity, direct: one key, one validation, replace the constant, a config
  test and a projects test.
- `tui-6159db` waits on the same answer; on a yes, file `Design column config from the
  brief` (high complexity) under `tui-cffb85` covering ordered subsets, the drop order,
  sort keys, and the `tui-0fa4d0` column.
- `tui-811ec5` waits on the pointer-input answer and `tui-c2e022`.
- On three "no"s, shelve `tui-6159db`, `tui-729559`, and `tui-811ec5` with wake
  conditions and close `tui-cffb85` with that verdict. (Taken.)
