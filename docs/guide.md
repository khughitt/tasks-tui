# Using tasks-tui

Type: guide; document types 1. How to move through tasks-tui's views, change a
task, launch an agent on one, and read what the app reports, as the current code
behaves. Installing, the key table and the configuration file are in the
[README](../README.md); what tasks-tui exchanges with the `tasks` CLI and a launched
harness is in [the tasks contract](tasks-contract.md).

## Views

Three views stack: Projects, Project and Task. `enter` or `i` opens the next one,
`esc` or `backspace` goes back, `q` quits, and the status line shows the stack as a
breadcrumb (`projects › <prefix> › <id>`). With no argument, tasks-tui opens the
Project view of the registered project whose root holds the current directory (the
deepest root wins), over Projects; elsewhere, or with `--all`, it opens Projects.
`tasks-tui <prefix>` opens that project and `tasks-tui <id>` opens that task over it.

**Projects** is a table with one row per registered project: an accent bar and the
prefix, the doing, todo, idea and blocked counts, and the age of the last
activity. The Name column is hidden; names still match the filter and appear in
Project details. Zero counts are dimmed, blocked stays red, and doing/todo/idea
use independent logarithmic blue–cyan–gold scales. A project the registry
cannot reach reads
`✗ unreachable` and does not open; a row reported without counts reads the same,
though it still opens. Rows start sorted by activity, newest first; `s p`/`S p` sort
by prefix and `s a`/`S a` by activity, ascending and descending. `/` filters on a
case-sensitive substring of the prefix or the name; `enter` or `esc` stops typing and
the filter stays applied until erased.

When the terminal is wide enough, a pane beside the table shows the highlighted
project's doing tasks, its parked tasks each followed by `→ <next step>`, and its
first eight ready tasks, with `N ready · 8 shown` when there are more. The pane loads
150 ms after the highlight stops moving. A strip under everything counts across all
projects: `3 doing · 2 parked across 19 projects`. The doing total sums the counts
returned by `projects`; projects without counts do not contribute to it.

**Project** opens with a header (accent bar, prefix, name, root, and doing, todo,
idea and blocked counts) over five tabs. The active tab shows how many rows it lists.

| Tab | Lists |
| --- | --- |
| 1 Ready | what the CLI's `ready` reports for the project |
| 2 Doing | tasks with status doing, merged with the project's parked tasks: a parked doing task stays one row carrying its park, and any other parked task is added as its own row |
| 3 Open | the CLI's default `list`: open tasks, by priority |
| 4 Ideas | tasks with status idea |
| 5 Done | tasks with status done, most recently updated first |

`1`–`5`, `tab`/`shift+tab` and `h`/`l` (or Left/Right) switch tabs; `j`/`k`,
`g g`/Home and `G`/End move. `/` filters case-insensitively on id, title and tags,
showing `/<text> · <shown> of <all>`. `s p`, `s a` and `s t` sort by priority, age
(newest first) and title; `S` reverses. The sorted column's header carries `↑` or `↓`,
rows without the value sort last, and the order survives reloads and tab switches.

A row shows the id in the project's accent, `P<n>`, size, complexity, process,
status, the age of the last update, a marker column, and the title followed by its
tags in brackets and the markers' words. Ages read `now`, then minutes (`m`), hours
(`h`), days (`d`) under two weeks and weeks (`w`); every valid age uses the
cyan-to-foreground recency scale. On a
narrow terminal the process column goes first, then complexity, size and age; the
title keeps at least 24 columns. The highlighted row carries an accent bar and bold text, with a faint
project tint only when TrueColor and a reported background are available.

| Marker | Meaning |
| --- | --- |
| `◆` and the owner | a live claim, in the accent |
| `◇` and the owner, dimmed | a stale claim: its session is gone |
| `⏸ user` or `⏸ agent` | parked, waiting on the user (yellow) or an agent (dim) |
| `⟳` and the cadence | a recurring task |
| `▸` and `N open` | a goal with open descendants |

A parked row the CLI could not resolve from here (for example, its task file exists
only on another branch) reads `unresolved` in the status column. It cannot be opened
(`enter` does nothing), and changing or launching it reports `no task highlighted`.

**Task** shows `show` in a scrolling viewport: the id and title; status, priority,
size, complexity, process, owner, created, started, updated, completed, tags and
source; the cadence and next due date of a recurring task; spec and plan with the
paths they resolve to and the plan step, flagged when the step is not found; the
checkout its commands run in; a claim (live or stale, owner, session, worktree); a
park (waiting on whom, reason, worktree, next step); and an escalation's level and
time. The body follows as rendered Markdown, then the notes in the order the CLI
gives them, then parent, children and dependencies with id, status and title; a
dependency that does not resolve from here reads `unresolved`. `j`/`k`,
`pgup`/`pgdn` and `g g`/Home scroll, and `y` copies the id through the terminal's
clipboard (OSC 52).

## Changing tasks

The transition keys act on the highlighted row in the Project view or the open task
in the Task view; in Projects they report `no task highlighted`. Park, done and drop
open a panel at the bottom, where `enter` submits and `esc` cancels. Each transition
runs one command in the task's checkout, then logs `<action> <id>` and reloads the
view, or shows the CLI's error.

- `space` **start** runs at once. When the CLI refuses because the task is claimed,
  the panel shows its message and `F` runs `start --force`; `esc` or `q` cancels. Any
  other refusal is shown and nothing is retried.
- `p` **park** asks for the one concrete next step, which is required. `ctrl+u`
  toggles waiting on an agent (the default) or the user; `ctrl+r` cycles the reason
  through none, review, decision, approval, environment, dependency and session.
  `capability` and `quiet` are not offered: they carry further arguments meant for
  agents (`--complexity`, and the `--minutes` that `quiet` requires), so park from
  the CLI for them.
- `d` **done** takes an optional message. A refusal (open descendants, open
  dependencies) is shown; `--force` is not offered.
- `x` **drop** takes an optional message, then asks for `y` to confirm, since dropping
  is one-way; `esc` or `q` cancels.

`a` opens the task form to add a task to the view's project (in Projects, the
highlighted one) and `e` opens it on the highlighted or open task, loaded with `show`
in its checkout and prefilled. The form has three fields: title (required), body
(several lines) and tags (space-separated, completed from the project's tags: Up/Down
choose a match, `enter` accepts it). `tab` and `shift+tab` move between fields,
`ctrl+s` saves, `esc` cancels; while a save runs, a second `ctrl+s` and `esc` do
nothing. An error shows under the fields until the next edit; a save closes the
form, logs what the CLI reports, and reloads the view beneath. Edit replaces the
title, body and the whole tag set. Add sets nothing else: no priority, size,
complexity or status (the CLI files a todo), and no agent, because a person filed it.

## Where commands run

Every command about one task (`show`, `edit`, `start`, `park`, `done`, `drop`) and
every launch runs in the task's checkout: the worktree its park names, else the
worktree its live claim names, else the project's registered root. A recorded
worktree that no longer exists falls through to the next choice with a warning that
names it, so a transition that lands on the root says why. Adding a task names the
project explicitly, so where tasks-tui was started never decides where it lands. The
rule is stated in full under [Checkout selection](tasks-contract.md#checkout-selection).

## Launching an agent

On the highlighted or open task, `c l`, `c o`, `c r` and `c O` launch claude, codex,
crush and opencode; `c c` opens a picker of every configured harness in name order
(`j`/`k` and `enter`, or a digit; `esc` or `q` cancels). A chord waits 900 ms for its
second key. A harness added under `[launch.harness.<name>]` in the config appears in the
picker; the four defaults are always configured.

A launch opens the configured terminal (kitty by default) in the task's checkout,
running the harness with a prompt naming the task, by default:

    Run `tasks start {id}` and continue that task: {title}

The harness gets tasks-tui's own environment less `TASKS_SESSION` and
`TASKS_SESSION_PID`, plus the harness's `env` table, and runs detached. The task's
status is untouched: the agent runs `tasks start` itself, so the claim belongs to the
session doing the work. The status line reports `launched
<harness> on <id> in <dir>`. A harness command without `{prompt}` (crush by default)
starts with no prompt; the status line says so and the prompt is copied to the
clipboard to paste. The exact argv and environment are under
[Launch handoff](tasks-contract.md#launch-handoff).

## Refresh

The view on top reloads after each write it issued, on `F5`, when it is opened or
returned to, and every `refresh_seconds` (30 by default; 0 turns the timer off).
Only the top view loads; results for views underneath are discarded and they reload
when they are on top again. The task form loads once and is never refreshed.

Reloads coalesce: while one is in flight, further requests leave one pending reload
that runs when it returns, so a reload after a write always reads after the write.
A result older than the latest request is discarded. A reload keeps the highlighted
task (or project) when it is still listed, and otherwise the same position; a tab
switch starts at the top. A spinner in the status line shows while the view loads.

## Messages and warnings

The status line's left side shows the latest message until the next key press: `✗`
errors in red, `⚠` warnings in yellow, `·` results in normal foreground. A less severe message
does not replace a more severe one still showing; with none, the breadcrumb shows.
Messages are outcomes: a write's result, each warning its response carried
(prefixed with the command and id), a failed load, write or launch, the
missing-worktree notice of a transition or launch. A failed load leaves the view's
previous rows in place.

A load's warnings describe what the view shows, so they are the view's data, not
messages: a reload replaces them. When the top view has any, `⚠ N` sits left of the
`? keys` hint. `W` opens the console: the top view's warnings under `warnings —
<view>`, then every message of the session, newest last (`j`/`k` scroll; `esc`, `W`,
`?` or `q` close). `?` opens the key legend; a box below its groups lists the sort
keys of the Projects and Project views. While a sort prefix (`s` or `S`) waits for its
second key, the `? keys` hint becomes that prefix, its direction (`↑` or `↓`) and the top
view's keys, e.g. `s … sort ↑  p priority  a age  t title`.

A failure at startup prints `tasks-tui: <message>` to stderr and exits 1 before the
screen opens; the conditions are listed under [Failures](tasks-contract.md#failures).

## Colours

Project colors come only from `tasks projects`: exact authored RGB in TrueColor,
the supplied `color_256` in indexed terminals, neutral in basic/no-color terminals.
Startup seeds the project map for direct-entry views; an accepted Projects reload
refreshes it. Uncolored projects remain neutral. Remove any old `[identity]`
config table when upgrading; familiar pins, scheme files and remote hashing are gone.

Ordinary labels, metadata, status/help text, form inputs and notices inherit the
terminal foreground. Inactive tabs, zero/absent values, stale claims and closed/
shelved work use terminal dim, with bold explicitly disabled. Doing status/title
is bold project accent; ideas are italic normal foreground. P0/errors/blocked use
ANSI red, P1/P2/warnings/user waits yellow (P1 bold), and recurrence cyan.

Task rows keep their selected gutter; Projects uses `▸ ` on the selected row and
`▌ ` otherwise. Active tabs and picker items use bold underlined project text;
uncolored active labels also have brackets. These indicators survive without a
background tint. TrueColor selection adds a faint mix of 12% project RGB into
the reported terminal background. Indexed/basic selections use no synthetic tint.

The app starts with an unknown profile and never assumes a dark theme. Only an
ANSI256/TrueColor profile requests foreground/background through Bubble Tea's
normal input loop; basic/no-color modes make no such queries. Late replies are
accepted, and a missing reply after one second yields at most one information
notice. Profile downgrades suppress stale notices. Until replies arrive,
background-dependent styles remain neutral.

Markdown uses normal foreground with bold/italic/underline emphasis, preserving
layout and wrapping without theme-specific syntax colors. Capable terminals do
not display literal inline Markdown delimiters; no-color modes retain them.


Counts share a blue–cyan–gold palette but normalize independently per column:
`log1p(n) / log1p(max(columnMaximum, 4))`. Maxima include all reachable projects
with counts before filtering, so scrolling/filtering cannot change a value's
color. A lone one stays moderate rather than maximum heat. A reload may change
normalization as the data changes. Exact counts remain visible.

Activity and task-table ages use a separate cyan-to-foreground recency fade:
`1 - log1p(ageMinutes) / log1p(43200)`, clamped from now to 30 days. Future dates
clamp to now; old dates stay at the normal foreground, missing activity is empty,
and invalid timestamps read `?`. The preview pane uses the same date scale.

Generated count colors meet 4.5:1 contrast against the actual row background;
date colors use the smaller of 4.5:1 and the terminal foreground's own contrast.
Authored project RGB and the terminal foreground remain exact. Indexed terminals
may quantize the generated data colors, so their final contrast depends on the
terminal palette. With only a background reply, counts can shade while dates
wait for the foreground; basic/no-color terminals show ordinary numeric text.

## Not built yet

Recorded direction, not current capability:

- Editing priority, size, complexity and process in the form, and adding and removing
  dependencies.
- Notes, block and unblock, shelve and unshelve from the TUI.
- Tree and dependency graph views.
- Showing lanes and needs from the tasks JSON.
