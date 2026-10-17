# Project identity brief

Updated 2026-10-17: ops moved identity into per-project files since the first pass; the user then settled the contract and placement.

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

- Answered by the user 2026-10-17: `tasks` exposes `name` and `purpose` in `tasks projects`, from each root's `identity.toml`, optional until adoption is complete (alternative 2).
- Answered by the user 2026-10-17: the purpose renders at the head of the selected project's pane.
- Open, for the tasks-side scoping: whether aliases travel too (the TUI's lean: filter-only, if at all), and how a malformed `identity.toml` surfaces.

## Proposed decomposition

- `tasks-6845e6` (idea in `tasks`): the `name` and `purpose` fields on `tasks projects` rows.
- `tui-0fa4d0` (todo, depends on `tasks-6845e6`): the TUI consumer — typed decoding, name column and filter from `name` with the basename when absent, the purpose line at the head of the pane, a recaptured `projects` fixture.
