# Task-list sorting — design

**Status:** draft. Goal: `tui-62bb3f`.

## Problem

Project tabs preserve the order returned by `tasks`, which is useful initially but
does not let a person inspect another visible task field without leaving the TUI.

## Decision

`S` in a loaded Project view opens a header selector. Left/right move through the
currently visible data columns (`id`, `P`, `sz`, `cx`, `proc`, `status`, `age`, and
`title`); gutter and marks are never selectable. Enter applies the selected column:
the first press sorts ascending, and pressing Enter on the active column reverses it.
Esc closes the selector without changing the sort. The active header shows an up or
down arrow; the selector gives its candidate header the existing selected surface.

The sort belongs to the Project view and therefore persists while switching tabs,
but no sort is active on the initial load, preserving the tracker's order. Sorting
is local: it never changes the `tasks` argv or reloads data. A resize can hide a
column; that column is unavailable to the selector but an active sort remains until
the person chooses another column.

Values sort by their corresponding row values: priority numerically; age by the
underlying `updated` timestamp; every other column lexicographically. Missing values
come after present values in either direction, and task id is the deterministic
tie-breaker. Filtering keeps its current behavior, then orders its matching rows by
the active sort while retaining the selected task when possible.

## Implementation

`projectView` owns the sort key, direction, and temporary header-selector state.
The shared table already computes visible widths, so the Project view derives the
selector's candidates from those widths rather than duplicating the responsive drop
rules. Its filter path performs the optional stable local sort after building the
filtered list. The table header accepts the active/suggested column decoration;
Projects-view sorting remains unchanged.

## Testing

Project-view tests cover selector navigation, ascending then descending priority,
timestamp ordering for `age`, lexicographic fields, missing values, selection
preservation across a filter, and unchanged task commands. Column tests cover that
dropped and structural columns cannot be selected.

## Out of scope

Mouse header clicks, configurable default sorts, multi-column sorts, and changing
the order supplied by `tasks` are not part of this task.
