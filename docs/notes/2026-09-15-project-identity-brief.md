# Project identity brief

## Problem

Use ops' canonical project names and purposes in the Projects view instead of deriving a name from each checkout path.

## Current behaviour and evidence

`internal/ui/projects.go` derives the displayed `name` column from `Project.Root`'s basename. `tasks projects` supplies prefix, root, reachability, counts, and activity only. `~/src/ops/identity.toml` maps each prefix to `name`, optional aliases, and `purpose`; it declares the tracker registry authoritative for project existence and location.

## Constraints

The TUI must continue to obtain project membership from `tasks`; it cannot silently fall back for missing named fields. The current table has no place for a purpose, and its narrow layout drops the `name` column. The identity file is owned by ops and its path, reload lifecycle, and schema contract are not yet a TUI interface.

## Alternatives

1. Read the file at a fixed external path and replace the name column. Rejected: that embeds a machine-specific cross-project path.
2. Add names and purposes to the `tasks projects` JSON contract. Lean: project metadata travels with the registry data and remains available to every client.
3. Add a TUI config path for the file. Rejected for now: duplicate registration and another source of drift.

## Unanswered questions

- Should `tasks projects` own this metadata, or should ops publish a stable reader/interface? Ops must answer.
- Where should a purpose render: a selected-project pane/header, task detail, or only wider layouts? The user must choose the desired density.
- Are aliases display/search metadata or only instructions for agents? Ops must answer.

## Proposed decomposition

- `tui-0fa4d0` remains an idea pending the metadata contract and presentation decision.
- If `tasks projects` gains metadata, create a bounded TUI consumer task covering typed decoding, table/pane rendering, and narrow-width behaviour.
