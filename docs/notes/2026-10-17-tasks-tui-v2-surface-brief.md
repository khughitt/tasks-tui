# tasks-tui v2 surface — brief

Ideas: tui-bbd26e (tree and dependency graph views), tui-d72add (field editing, dep,
shelve/unshelve, note, block), tui-c6e459 (inotify instead of the refresh timer),
tui-f67715 (lanes, in_lane, needs). The first three are the v1 deferrals of
docs/specs/2026-09-13-tasks-tui-v1-design.md §14; the fourth arrived with the tasks
lanes contract.

## 1. Problem

v1 is the daily surface: look, read, file, launch, and start/park/done/drop. A person
still drops to the CLI to see how a project's work hangs together (hierarchy and
dependencies), to change anything but title, body and tags, and to see which efforts
can run now. The view also lags agents by up to `refresh_seconds` (30 s by default).

## 2. Current behaviour and evidence

- Task view relations block lists parent, children and dependencies from `show`
  (§5.3). `show` has no dependents (tasks-e8d471 is open for that).
- `tasks tree [--project p | --all-projects] [--all]` returns nested `nodes` in the
  list-row shape. `tasks graph` emits only mermaid or dot text in a JSON wrapper, has
  no `--project`, and is not structured data. List rows already carry `depends`, so a
  per-project edge set is available from `list --project p` without a new contract.
- Every write d72add names exists in the CLI: `edit -p/--size/--complexity/--process/
  --every/--defer/--parent/--need`, `dep <id> --on/--rm`, `note`, `block`, `unblock`,
  `shelve <id> <wake>`, `unshelve`. The TUI wires only `edit --title -b --tag`
  (internal/tasksctl/client.go).
- Refresh is a `tea.Tick` per `refresh_seconds` (internal/ui/app.go). Claims and parks,
  the main reason to refresh, live in the out-of-git claim store
  (`$XDG_STATE_HOME/tasks/claims/<prefix>.toml`), not under `tasks/`; tasks owns that
  path. Measured 2026-10-17: `tasks projects` 0.19 s, `prime --project tui` 0.01 s,
  `list --all-projects --status doing` 1.82 s (the Projects strip runs it and `--parked`
  on every tick).
- Lane data is live: list rows carry `lane`, `in_lane`, `needs`; `tasks lanes
  --all-projects` returned 36 lanes with `guidance`, `state`, `pick`, `steps`, `active`.

## 3. Constraints

- The TUI never reads `tasks/*.md` or the claim store; anything it needs and cannot
  get is a tasks task (AGENTS.md, v1 §2, §15).
- Every new binding is a row in ops's tools/keys.toml, vendored byte-identical and
  checked by the keys test; the change lands in ops first, then is re-vendored.
- The task form is specified as exactly three fields (v1 §6); widening it revises that.
- tui-6159db (columns from config) is shelved; the column model (internal/ui/columns.go)
  is where new columns go.

## 4. Alternatives

Tree and graph (bbd26e):
- A. A **Tree** tab in the Project view from `tasks tree --project`, indented rows with
  the ordinary columns; dependencies stay in the Task view, gaining dependents once
  tasks-e8d471 lands. *Current lean.*
- B. A and a per-task **neighbourhood** view: the task, what it depends on, and what
  depends on it, built from list-row `depends`, walkable with `enter`.
- C. A whole-project graph drawn in the terminal. Layout of a general DAG in a grid is
  the cost; it is the least readable at the sizes seen (30+ open tasks).

Refresh (c6e459):
- A. Keep the timer; make the Projects strip slower than the focused view. *Current
  lean, unless the lag is felt.*
- B. A tasks-side change feed (`tasks watch` streaming change events, or a cheap
  change stamp per project) that the TUI consumes; filed in tasks if wanted.
- C. The TUI watches `tasks/` directories itself: misses claims and parks, and the
  Projects view would watch every registered root and worktree. Rejected.

## 5. Unanswered questions

- Which of tree A/B/C is wanted, and is a whole-project graph needed at all? The
  design task (tui-c30d88) frames it for the user; the user decides.
- Is the 30 s lag or the poll cost felt in use? Only the user can say (on c6e459).

## 6. Proposed decomposition

- Goal tui-29c34d holds the cluster.
- tui-c30d88 — Design the tree and dependency views from this brief; tui-bbd26e waits
  on it.
- tui-d72add, now a goal: tui-55cb9e (note, block/unblock, shelve/unshelve overlays,
  direct), tui-662270 (rubric fields in the task form, planned), tui-c2a9df
  (dependency editing, planned).
- tui-f67715 scoped (planned): lane data in rows and a lanes pane.
- tui-c6e459 waits on the user's answer.
