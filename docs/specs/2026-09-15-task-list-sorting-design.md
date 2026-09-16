# Task-list sorting — design

**Status:** approved 2026-09-16. Goal: `tui-62bb3f`.

## Problem

Project tabs preserve the order returned by `tasks`, which is useful initially but
does not let a person inspect another visible task field without leaving the TUI.

## Decision

`S` in a loaded Project view opens a capturing header selector. Left/right move
through the currently visible data columns (`id`, `P`, `sz`, `cx`, `proc`, `status`,
`age`, and `title`); gutter and marks are never selectable. It starts at the active
column when that column is visible, otherwise at the first data column. Enter applies
the candidate and closes the selector: the first press sorts ascending, the second
press on that column reverses it, and the third clears the sort. Esc and backspace
close the selector without changing the sort. Other keys are ignored. Filtering and
the selector are mutually exclusive, so `S` remains filter text while filtering.

The active header shows an up or down arrow; the selector gives its candidate header
the existing selected surface. Decorated labels participate in table measurement, so
an indicator neither clips nor shifts a column.

The sort belongs to the Project view and therefore persists while switching tabs,
but no sort is active on the initial load, preserving the tracker's order. Sorting
is local: it never changes the `tasks` argv or reloads data. Reloads and tab switches
apply an active sort to their fresh rows; it overrides the Done tab's server ordering.
A resize can hide a column; its indicator is hidden too, but its ordering still
applies. Selecting a visible column replaces that sort, and continuing that column's
three-state cycle eventually clears it.

Values sort by their corresponding row values: priority numerically; size by
`xs < s < m < l < xl`; complexity by `low < mid < high`; status by
`idea < todo < doing < blocked < shelved < done < dropped`; and age by the underlying
`updated` timestamp. Id and process sort lexicographically; title is
case-insensitive lexicographically. Unparseable age and unresolved parked status are
missing values. Missing values come after present values in either direction, and
task id is the deterministic tie-breaker. Filtering keeps its current behavior, then
orders its matching rows by the active sort while retaining the selected task when
possible.

## Implementation

`projectView` owns the sort key, direction, and temporary header-selector state; its
capture predicate includes that state so App routes Esc and backspace to the view.
The shared table already computes visible widths, so the Project view derives the
selector's candidates from those widths rather than duplicating the responsive drop
rules. Its filter path performs the optional stable local sort after building the
filtered list. The table header accepts the active/suggested column decoration;
Projects-view sorting remains unchanged.

## Testing

Project-view tests cover selector navigation and capture, the three-state priority
cycle, timestamp ordering for `age`, ranked enum and lexical fields, missing values,
selection preservation across a filter, reload and tab persistence, and unchanged
task commands. They also prove Esc/backspace close the selector without popping the
view. Column tests cover that dropped and structural columns cannot be selected and
that decorated headers measure wide enough. The global `S` help changes from
`sort projects` to `sort`, with the v1.1 legend table updated accordingly.

## Out of scope

Mouse header clicks, configurable default sorts, multi-column sorts, and changing
the order supplied by `tasks` are not part of this task.
