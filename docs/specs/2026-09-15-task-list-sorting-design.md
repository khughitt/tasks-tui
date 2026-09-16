# Task-list sorting — design

**Status:** sorting implemented 2026-09-16 under `tui-62bb3f`; interaction updated
to the shared chord vocabulary by `tui-8d070b` on 2026-09-16.

## Problem

Project tabs preserve the order returned by `tasks`, which is useful initially but
does not let a person inspect another visible task field without leaving the TUI.

## Decision

`s p`, `s a`, and `s t` sort a Project view by priority, age, and title ascending.
`S p`, `S a`, and `S t` select descending. Repeating a chord keeps that direction.
The shared key vocabulary supersedes the original `S` header selector. App holds
one prefix for 900 ms and shows it in the status line. Filters and overlays own
text input, including these prefix characters.

The active header shows an up or down arrow. Decorated labels participate in table
measurement, so an indicator neither clips nor shifts a column. Sorting belongs to
the Project view and persists while switching tabs and reloading; the initial view
preserves tracker order. Sorting is local and does not change `tasks` argv.

Priority is numeric; ascending age puts the youngest updated timestamp first;
titles compare case-insensitively. The landed sorting helpers retain missing
values last in either direction and task ID as a deterministic tie-breaker.
Filtering preserves the selected task when possible. A newly loaded tab starts at
its first row; a refresh of the same tab preserves selection.

Projects uses `s p`/`S p` for prefix and `s a`/`S a` for activity, defaulting to
activity descending. Its table retains the selected project after sorting.

## Implementation

`projectView` retains `sortKey`, `descending`, and its existing comparison helpers.
The chord table sets those fields directly, and `applyFilter` applies the stable
local sort. `capturing()` now covers filtering only. Header arrows use the existing
measured-label path. The earlier `cycleSort` helper remains, but no user key cycles
sorting off; a new view starts in tracker order.

## Testing

Chord tests cover priority/title direction, youngest-first age, header arrows,
idempotent Projects sorting, persistent sorting across tabs, and navigation.
Existing sort-helper tests retain selection and missing-value coverage. The
conformance test compares every scoped binding and chord argument against the
vendored inventory; the legend tests check the displayed key labels.

## Out of scope

Mouse header clicks, configurable default sorts, multi-column sorts, and changing
the order supplied by `tasks` are not part of this task.
