# tasks-tui v1: the daily surface — design

**Status:** draft 2026-09-13, awaiting review. Goal: tui-d6e352. Plan: none yet.

## 1. Problem

`tasks` is a CLI with a JSON contract and a `--pretty` mode, and both are built for
agents and for one command at a time. A person looking across nineteen registered
projects has no place to see them all, drop into one, read a task, file a thought in
one line, or hand a task to a coding agent, without composing several commands and
reading their output in a scrollback.

tasks-tui is that place: a keyboard-driven terminal application over the tracker.
It shows every project and one project, shows one task in full, files a task from a
single line with inline tokens, launches an agent session on a task in a new terminal
window, and performs the four transitions a person makes during a day — start, park,
done, drop. Everything else stays in the CLI.

## 2. Decisions

- **A separate repository, in Go, on Bubble Tea.** `tasks-tui` lives beside `tasks`
  as its own registered project (`tui`). Rejected: a directory inside the tasks
  repository — a second toolchain under the Rust gates, and a consumer sitting beside
  the contract it consumes, where every JSON change would tempt a lockstep edit.
- **The `tasks` binary is the only reader and the only writer.** The TUI runs
  `tasks` as a subprocess and parses its JSON. It never opens `tasks/*.md`, the
  registry, or the claim store. Rejected: a Go parser of the task files — a fork of
  the format that drifts, and a violation of the tracker's one-writer rule the moment
  it learns to write.
- **The JSON contract is consumed as documented, never inferred.** Every field the TUI
  reads is named in §4. Unknown fields are ignored; a missing named field is a typed
  error, not a zero value.
- **A subprocess is run with an explicit environment.** `TASKS_FORMAT`, `TASKS_COLOR`,
  `TASKS_AGENT`, and `TASKS_MAX_COMPLEXITY` are removed from every child's
  environment: the TUI needs JSON, is operated by a person rather than an agent, and
  must show the whole ready list. `TASKS_SESSION=tasks-tui:<pid>` and
  `TASKS_SESSION_PID=<pid>` are set so a claim made from the TUI is attributable and
  its liveness follows the TUI process.
- **Launch never changes a task's status.** The launched agent runs `tasks start`
  itself, so ownership and the claim belong to the session doing the work, not to the
  window that opened it. Rejected: `start` on launch — the TUI would hold a live
  claim the agent then has to `--force` past.
- **A project's accent is familiar's.** The same slot a project has in familiar — a
  pin in `~/.config/familiar/identities.yaml`, else `fnv1a32(key) mod 12` over the
  same key familiar hashes — selects the same frozen hue table. The algorithm is
  reimplemented; nothing from familiar is imported at build or run time, which is the
  boundary ops-df0842 asks for. When ops ships a visual identity, the TUI reads it
  instead. Rejected: a palette keyed by prefix — a project would then look different
  in the TUI than on its pet and its terminal glass.
- **v1 is the daily surface.** Views, detail, quick add, quick launch, and
  start/park/done/drop. Field editing, dependencies, shelving, notes, and the tree
  are out (§14).
- **Fail early.** A `tasks` that is missing or does not answer `projects`, a malformed config, an unregistered
  project in a quick-add token, or a CLI error all surface as a message; nothing is
  retried, defaulted, or hidden.

## 3. Architecture

One binary, `tasks-tui`, Go module `tasks-tui`. Bubble Tea v2 (`bubbletea/v2`
2.0.x), Lip Gloss v2, Bubbles v2, Glamour v2 for the task body. Config in TOML
(`BurntSushi/toml`); familiar's pins in YAML (`gopkg.in/yaml.v3`).

```
cmd/tasks-tui/main.go        flags, config load, tasks probe, tea.NewProgram
internal/tasksctl/           the subprocess boundary: Runner, typed calls, JSON types, Error
internal/identity/           familiar's key, pins, fnv1a32, slot hues, tone → lipgloss colors
internal/quickadd/           the one-line grammar: Parse(line) → Spec | error; Spec.Args()
internal/launch/             harness config, prompt template, terminal spawn
internal/ui/                 the tea model: app, views, overlays, keymap, styles
internal/ui/projects/        the Projects view
internal/ui/project/         the Project view
internal/ui/task/            the Task view
internal/ui/overlay/         prompt, picker, confirm
```

Data flows one way. A view asks the app for data; the app issues a `tea.Cmd` that
calls `tasksctl`; the result comes back as a message carrying either typed data or a
`tasksctl.Error`; the view re-renders. Views hold no subprocess handles and never
call `tasksctl` directly. A write command (start/park/done/drop/add) is followed by a
reload of the view that issued it; nothing is patched optimistically.

Each package answers three questions: what it does, how it is used, what it depends
on. `tasksctl` depends on `os/exec` and the JSON types. `identity` depends on the
filesystem (pins, scheme, git remote). `quickadd` depends on nothing but the token
grammar and the set of registered prefixes it is given. `launch` depends on
`os/exec`. `ui` depends on all four and on the charm libraries.

## 4. The tasks contract as used

`tasks` reports `0.1.0` and has never bumped it, so a version floor would say
nothing. At startup the TUI runs `tasks projects`; a missing binary, a non-zero exit,
or a response without the fields below is a fatal message naming what was expected.
Drift after that surfaces the same way: a named field missing from any response is a
typed decode error shown in the status line, never a zero value.

Every invocation is `tasks [-C <dir>] <command> [flags]` with the environment of §2,
stdout parsed as JSON, and a non-zero exit parsed as `{"error": {"kind", "detail"}}`
into `tasksctl.Error{Kind, Detail}`. A non-zero exit whose stdout is not that shape
is `Error{Kind: "unparseable", Detail: <stderr>}`.

Reads, and the fields consumed:

| Call | Fields used |
|---|---|
| `projects` | `projects[]`: `prefix`, `root`, `reachable`, `counts.{idea,todo,doing,blocked}`, `last_activity` |
| `prime --project <p>` | `doing[]`, `parked[]`, `ready[]`, `counts`; rows are list rows |
| `ready --project <p>` | `tasks[]` list rows |
| `list --project <p> [--status …] [--sort updated]` | `tasks[]` list rows, `warnings[]` |
| `list --all-projects --status doing` and `--parked` | list rows across projects |
| `show <id>` | `task.{id,title,status,priority,size,complexity,process,every,owner,created,updated,started,completed,tags,source,spec,plan,step,body,notes[]}`, `spec_path`, `plan_path`, `depends_on[]`, `parent`, `children[]`, `claim`, `park`, `escalation`, `periodic` |
| `root <id>` | the registered root of the task's project |

A list row's consumed fields: `id, title, status, priority, size, complexity,
process, owner, updated, tags, parent, child_count, open_descendant_count, claim,
park, periodic`. Row `claim` gives `owner, session, worktree`; row `park` gives
`next_step, waiting_on, reason, worktree, at`.

Writes:

| Action | Command |
|---|---|
| add | `add "<title>" --project <p> [--status idea] [-p N] [--size S] [--complexity C] [--every Nd] [--tag t]… [-b "<body>"]` |
| start | `start <id>`; on `Kind == "claimed"` the TUI offers `start --force <id>` |
| park | `park <id> "<next step>" [--waiting-on user] [--reason R]` |
| done | `done <id> ["<message>"]` |
| drop | `drop <id> ["<message>"]` |

`--project` is always passed on `add`; the TUI's own working directory never decides
where a task lands. Write commands run with `-C <registered root>` so `owner` records
the branch of the main checkout, as a person at that checkout would get.

## 5. Views and keys

Three views on a stack: Projects → Project → Task. `esc` and `backspace` pop;
`q` quits from anywhere outside an overlay; `?` toggles a key legend; `r` reloads
the current view. Started inside a registered root (`tasks projects` says so), the
app opens that Project view with Projects beneath it; started elsewhere, or with
`--all`, it opens Projects. `tasks-tui <prefix>` opens that project; `tasks-tui
<id>` opens that task.

### 5.1 Projects

A table of every registered project, one row each: accent bar, prefix, name (the
root's basename), `doing todo idea blocked` counts, last activity as a relative
time, and a dimmed row with `✗` when unreachable. Default order is last activity,
newest first; `s` toggles prefix order. `/` filters rows by prefix or name.

To the right, a pane for the highlighted project from `prime --project <p>`:
`doing` rows with their claim owner, `parked` rows with waiting-on and the next step,
and the first eight `ready` rows. It loads on a 150 ms debounce after the highlight
moves. `enter` opens the project. `a` quick-adds into the highlighted project.

Below the table, a strip: `list --all-projects --status doing` and `--parked`
counted per project — "3 doing · 2 parked across 19 projects" — so the person sees
in one line whether anything is waiting on them.

### 5.2 Project

A header with the accent bar, prefix, name, root, and counts. Under it a tabbed
list: **Ready** (default; `ready --project`), **Doing** (`list --status doing`
followed by `list --parked`, parked rows tagged), **Open** (`list`, the default
statuses by priority), **Ideas** (`list --status idea`), **Done** (`list --status
done --sort updated`, most recent first). `1`–`5` and `tab`/`shift+tab` switch tabs;
`/` filters by id, title, or tag substring.

A row: `id  P<n>  size  complexity  process  status  updated  title  [tags]`, with
a claim marker (`◆ owner`) and a park marker (`⏸ user` or `⏸ agent`) after the
title when present, `⟳ every` for a recurrence, and `▸ n` for a goal with open
descendants. `enter` opens the task; the transition keys of §5.4 act on the
highlighted row without opening it; `a` quick-adds into this project; `l` launches
the highlighted task.

### 5.3 Task

`show <id>` rendered in a scrollable viewport: a field block (id, status, priority,
size, complexity, process, owner, created/started/updated/completed, tags, source,
every and next due, spec and plan with their resolved paths and `step`), the body
through Glamour with a style matching the tone (§8), notes newest last, and a
relations block: parent, children, and dependencies each with id, status, and
title. A live claim shows owner, session, worktree; a park shows next step,
waiting-on, reason, and the checkout it was parked in; an escalation shows its level.

Keys: the transitions of §5.4, `l` launch, `y` copies the id to the clipboard when
`wl-copy` is on `PATH` (otherwise the key reports that it is not), `j/k` and page
keys scroll.

### 5.4 Transitions

Each key opens an overlay at the bottom of the screen; `esc` cancels it; `enter`
submits. Every submit runs one command, shows its `Error.Detail` on failure, and
reloads the view on success.

- `s` **start** — no prompt. On `claimed`, the overlay shows the claim's owner and
  session and offers `F` to run `start --force`.
- `p` **park** — a required one-line next step; `ctrl+u` toggles waiting-on
  between agent and user; `ctrl+r` cycles the reason through none, review, decision,
  approval, environment, dependency, session. `capability` and `quiet` are not
  offered: both take arguments (`--complexity`, `--minutes`) the CLI validates and
  the skill reserves for agents; the CLI is the place for them.
- `d` **done** — an optional message. A refusal (open descendants, open
  dependencies) is shown; `--force` is not offered.
- `x` **drop** — an optional message, with a confirm since dropping is a
  one-way status.

## 6. Quick add

`a` opens a single-line input at the bottom of any view. The line is parsed on
every keystroke; tokens are colored in place, and a preview line under the input
shows the command that would run, or the first error in the error color. `enter`
files it when there is no error and is refused otherwise; `esc` cancels.

### 6.1 Grammar

The line is split on whitespace. A word is a token when it *starts* with a marker
and the rest matches; a word that starts with a marker and does not match is an
error, never title text. Everything that is not a token is title text, joined with
single spaces in order.

| Token | Field | Rule |
|---|---|---|
| `#<tag>` | `--tag` | repeatable; tag matches `[A-Za-z0-9_:./-]+` |
| `!<n>` | `-p` | `n` in 0–4 |
| `~<size>` | `--size` | xs, s, m, l, xl |
| `^<level>` | `--complexity` | low, mid, high |
| `@<n>d` / `@<n>w` | `--every` | positive integer |
| `><prefix>` | `--project` | must be a registered prefix |
| `?` | `--status idea` | only as the first word, alone or attached (`?fix …`) |
| ` -- ` | `-b` | the rest of the line, verbatim, is the body |

A duplicate single-valued token (`!2 … !3`) is an error. An empty title is an
error. The project is the `>` token when present, else the view's project (Projects:
the highlighted row; Project and Task: that project). Nothing else is defaulted:
no size, priority, or complexity is invented, and no `--agent` is passed, so a task
filed here carries `agent: null`, which is what a person filing means.

Example: `?tint the projects strip with each accent #ui #identity !3 ~s`
→ `add "tint the projects strip with each accent" --project tui --status idea
--tag ui --tag identity -p 3 --size s`.

### 6.2 Styling

Each token class has a color: tags in the project accent, priority in the priority
color of §8, size and complexity in the muted foreground, recurrence in the periodic
color, `>prefix` in that project's accent, `?` and the body separator dimmed. The
title stays in the plain foreground. The preview line uses the same colors on the
rendered flags.

## 7. Quick launch

`l` on a task opens a picker of configured harnesses. Choosing one spawns a terminal
window running that harness in the task's checkout with an initial prompt, detaches
it, and reports "launched <harness> on <id> in <dir>" in the status line. The task's
status is untouched (§2).

**Checkout:** the park record's `worktree` when the task is parked and that path
exists, else the registered root from `root <id>`. A parked worktree that no longer
exists is an error naming it, not a fallback.

**Command:** `terminal` with `{dir}` substituted, then the harness `command` with
`{prompt}` substituted, as one argv. A harness command without `{prompt}` runs
without one and the status line says so. The process is started in its own session
(`Setsid`), stdio to `/dev/null`, environment inherited minus `TASKS_SESSION`,
`TASKS_SESSION_PID`, and plus the harness's `env`. Spawn failure is an error;
what the harness does afterwards is not the TUI's business.

**Prompt:** `prompt` with `{id}` and `{title}` substituted. The default is
`Run \`tasks start {id}\` and continue that task: {title}` — enough for a harness
that loads the tasks skill to find the record and follow its process.

**Defaults** (all overridable in config):

```toml
[launch]
terminal = ["kitty", "--directory", "{dir}", "--"]
prompt   = "Run `tasks start {id}` and continue that task: {title}"

[launch.harness.claude]
command = ["claude", "{prompt}"]

[launch.harness.codex]
command = ["codex", "{prompt}"]

[launch.harness.opencode]
command = ["opencode", "--prompt", "{prompt}"]

[launch.harness.crush]
command = ["crush"]            # interactive crush takes no initial prompt
```

Verified 2026-09-13 against the installed binaries: `claude` and `codex` take the
prompt as a positional, `opencode --prompt` exists, interactive `crush` has no
prompt flag. A shell function that wraps a harness (the `crush` and `opencode`
wrappers exporting `TASKS_MAX_COMPLEXITY`) is not seen by a direct spawn; the
harness's `env` table carries it: `env = { TASKS_MAX_COMPLEXITY = "mid" }`.

This subsumes tasks-202e1f; when the plan lands, that idea gets a note and is
dropped in favor of the TUI.

## 8. Identity and style

### 8.1 A project's slot

`identity.SlotFor(root)`:

1. Canonicalize `root` with `filepath.EvalSymlinks`.
2. Read `~/.config/familiar/identities.yaml` if present: `identities[].{path, slot}`.
   Each `path` is `~`-expanded and canonicalized; the first whose canonical path
   equals the root's wins. A file that exists but does not parse is an error.
3. Otherwise `key = remote ?? root`, where `remote` is
   `git -C root remote get-url origin` normalized exactly as familiar does
   (scheme, credentials, port, `.git` stripped; scp form folded; lowercased;
   `null` when it is not `host/owner/name`) and `root` is the canonical path.
   `slot = fnv1a32(key) mod 12`.

The twelve `(hue, sat)` pairs are copied from familiar's `slot-hues.js` and marked
frozen with the same warning. The test suite carries familiar's `normalizeRemote`
cases and a set of `fnv1a32` vectors so a drift in either shows up here.

### 8.2 Tone

`~/.config/familiar/scheme.json` gives `mode` (dark or light) and `satScale`;
absent, the tone is dark with scale 1. familiar's `ramp` lightness anchors are
reused: the accent is `(hue, sat × satScale, base)` with base 58 in dark and 46 in
light; a dim variant uses `shadow`; the Glamour style is `dark` or `light` by mode.

### 8.3 Semantic colors

Independent of the project accent, and the same in every project:

| Meaning | Use |
|---|---|
| priority | P0 red, P1 orange, P2 yellow, P3 default foreground, P4 dim |
| status | todo default, doing accent, blocked red, idea dim italic, done/dropped dim strikethrough |
| park waiting on user | yellow marker; waiting on agent, dim marker |
| claim | accent marker |
| periodic | teal marker |
| error | red, in the status line and the quick-add preview |

Each is a light/dark pair resolved once from the tone (`lipgloss.LightDark`), chosen
from the terminal's 256-color range so it sits inside any scheme. They are the only
hard-coded colors in the program; everything project-specific derives from the slot.

## 9. Config

`$XDG_CONFIG_HOME/tasks-tui/config.toml`, default `~/.config/tasks-tui/config.toml`.
Absent is fine; present and malformed is a fatal message naming the key.

```toml
tasks = "tasks"                  # binary name or path
refresh_seconds = 30             # 0 disables the timer

[launch]                         # §7
terminal = [...]
prompt = "..."

[launch.harness.<name>]
command = [...]
env = { KEY = "value" }

[identity]
slot = { tui = 2 }               # per-prefix override, wins over familiar's pins
```

Unknown keys are errors. The `identity.slot` override exists so a project without a
familiar pin can be placed deliberately; when ops-df0842 ships a shared identity,
this table is the first thing to delete.

## 10. Refresh

The current view reloads after any write it issued, on `r`, and on a timer every
`refresh_seconds` (agents change claims and parks under the TUI). A reload preserves
the highlighted id when it is still present and the scroll position otherwise. A
reload in flight when another is requested is dropped, not queued. The Projects
right pane is debounced as in §5.1.

## 11. Errors

Every failure has one home: the status line at the bottom, in the error color,
until the next keypress. It shows `Kind: Detail` for a `tasksctl.Error`, the
spawn error for a launch, the parse error for quick add. A fatal condition at
startup (no `tasks`, version too low, malformed config, unparseable pins) prints the
message to stderr and exits 1 before the alternate screen opens. Nothing is logged
to a file; nothing is retried.

## 12. Testing

- `quickadd`: table tests over the grammar — every token, every error, ordering,
  the body separator, the example of §6.1 producing that exact argv.
- `identity`: familiar's `normalizeRemote` cases ported verbatim; `fnv1a32` vectors
  (empty string, `a`, a known remote key) with expected values computed from
  familiar's implementation; pin matching through a symlinked path; the slot table
  length asserted against 12.
- `tasksctl`: the argv each typed call builds; JSON fixtures captured from the real
  binary under `testdata/` decoded into the types; the error envelope; the
  environment scrubbing of §2, asserted on the `exec.Cmd` before it runs.
- `launch`: argv assembly for each default harness, `{prompt}` absent, checkout
  selection with and without a park, the missing-worktree error.
- `ui`: `teatest/v2` runs over a `tasksctl.Runner` interface faked in memory —
  open a project, switch tabs, filter, open a task, park with a next step and
  waiting-on user, the argv the fake received; quick add end to end from keystrokes
  to argv.
- One integration test, skipped when `tasks` is not on `PATH`, that inits a scratch
  project with `XDG_CONFIG_HOME` pointed at a temp dir, adds two tasks through the
  real binary, and drives the TUI's project view against it.

## 13. Repository layout and gates

```
cmd/tasks-tui/           internal/…            docs/specs/  docs/plans/
tasks/                   justfile              tools/tt     .githooks/
```

`just check`: `gofmt -l`, `go vet ./...`, `staticcheck ./...`, `tasks check`.
`just test`: `go test ./...`. `just gate`: both. `just setup`: `go mod download`
and `git config core.hooksPath .githooks`. Every recipe runs through `tools/tt`, a
byte-identical copy of ops `bin/tt`, so the runs join the cross-project audit. The
hooks run `check` at pre-commit and `gate` at pre-push. `just install` is
`go install ./cmd/tasks-tui`.

Conventional commits, no AI-attribution trailers. An `AGENTS.md` adopting the tasks
process policy is part of the first commit of the plan.

## 14. Out of scope for v1

Filed as ideas in `tui` when the plan lands, not built now:

- Field editing (`edit`), dependencies (`dep`), shelve/unshelve, `note`, `block`.
- The tree and dependency graph views.
- Tag completion in quick add from `tasks tags`.
- A clipboard fallback for harnesses without an initial prompt.
- Reading ops `identity.toml` for project names and purposes; the basename serves
  until visual identity (ops-df0842) decides where identity lives.
- Watching `tasks/` with inotify instead of the timer.

## 15. Cross-project effects

- `tasks`: no change. The TUI consumes the contract as it stands; anything it turns
  out to need is filed there as a task, never patched around.
- `tasks-202e1f` (quick launch idea): noted and dropped when §7 lands.
- `ops-df0842` (visual identity): this design is a consumer; §8.1 and §9 name
  the seams that change when it ships.
- `familiar`: read-only use of two config files; no code shared.
