# Project identity brief

Updated 2026-10-17: ops moved identity into per-project files since the first pass.

## Problem

Use the canonical project names and purposes in the Projects view instead of deriving a name from each checkout path.

## Current behaviour and evidence

`internal/ui/projects.go` derives the displayed `name` column (and the filter match) from `Project.Root`'s basename (`name()`). `tasks projects` (tasks 0.4.0) supplies prefix, colour, root, reachability, counts, total, and last activity only; it carries no name or purpose.

Since 2026-09-22 (ops-03f3e3, ops's project-guides design §6) identity has two homes. Each adopting project holds a root `identity.toml` (`name`, `aliases`, `purpose` of at most 80 characters, `abstract`, and more). Ops keeps `identity-mirror.toml`, one table per prefix, which `ops-projects pull` refreshes from the reachable roots. On 2026-10-17, 16 of 34 registered roots hold an `identity.toml`; tasks-tui does not, but the mirror has a `[tui]` table. The tasks project has no open task to expose identity in `tasks projects`.

## Constraints

The TUI must continue to obtain project membership from `tasks`. AGENTS.md: anything the TUI needs from the JSON contract and cannot get is a task in `tasks`, never a workaround here. Missing required fields are decode errors; nothing falls back silently. The current table has no place for a purpose, and its narrow layout drops the `name` column first. Identity files are ops's schema, not a TUI interface.

## Alternatives

1. Read the ops mirror at a fixed path. Rejected: a machine-specific cross-project path.
2. `tasks projects` gains `name` and `purpose` (and possibly `aliases`), read from each root's `identity.toml` or the mirror. Lean: the metadata travels with the registry data every client already reads, and the TUI change is a typed decode plus rendering.
3. The TUI reads `<root>/identity.toml` for each registry root itself. Now path-free, since roots come from the registry, but it is half-adopted (16/34), so most rows would still need the basename, and it is a second contract outside `tasks`.
4. A TUI config path for the file. Rejected: duplicate registration and another source of drift.

## Unanswered questions

- Should `tasks` expose name and purpose in `tasks projects` (alternative 2), and from which source: per-project `identity.toml` or the mirror? The user, as owner of tasks and ops, decides. Recommendation: yes, from the root's `identity.toml`, with the field optional until adoption is complete.
- Where should a purpose render: the selected-project pane header, a wide-layout column, or task detail? The user's call on density. Lean: the pane header, which costs no column.
- Are aliases display or filter metadata, or only instructions for agents? Lean: filter-only, if at all.

## Proposed decomposition

- `tui-0fa4d0` stays an idea, awaiting the user on the first two questions.
- If alternative 2 is chosen: a task in `tasks` for the fields, then a bounded TUI consumer task here (typed decoding, name column and filter from the new field, purpose placement, narrow-width behaviour, a recaptured `projects` fixture) that depends on it.
