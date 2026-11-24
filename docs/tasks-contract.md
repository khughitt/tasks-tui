# The tasks contract

Type: protocol; document types 1. Currently unversioned. What tasks-tui sends to the
`tasks` CLI and reads back, in order: invocation and environment, the startup probe,
the checkout a task's commands run in, every read and write with the fields it uses,
the handoff to a launched harness, failures and known gaps. What a person sees of
it is in [the guide](guide.md); installing and configuring are in the
[README](../README.md).

## Participants

- **tasks-tui**, the terminal app.
- **The `tasks` CLI** ([tasks](https://github.com/khughitt/tasks)), the only reader
  and writer of task data. tasks-tui never opens task files, the registry or the
  claim store: every read and write is one `tasks` subprocess whose JSON it decodes.
- **A terminal and a harness**, started for a launch (kitty and claude, codex,
  crush or opencode by default). tasks-tui hands them an argv, a directory and an
  environment, and never talks to them again.

## Invocation and environment

Every call is `tasks [-C <dir>] <command> [args…]`, with the binary named by the
config's `tasks` key (`tasks` by default, found on `PATH`). `-C <checkout>` is
passed only for a command about one task (see
[Checkout selection](#checkout-selection)); every other call omits it. A call in the
running app is killed after 10 s; at startup, the probe and the lookups for the
opening view share one 20 s limit. No call passes `--json`: JSON is the CLI's default.

The child's environment is tasks-tui's own with these variables removed:

| Removed | Why |
| --- | --- |
| `TASKS_FORMAT`, `TASKS_COLOR` | the output must be the default, uncoloured JSON |
| `TASKS_AGENT`, `TASKS_MODEL` | a person operates the app, so no write may stamp a harness or model |
| `TASKS_MAX_COMPLEXITY` | the ready list must be whole, not cut to an agent's cutoff |
| `TASKS_SESSION`, `TASKS_SESSION_PID` | replaced below |

and these added: `TASKS_SESSION=tasks-tui:<pid>` and `TASKS_SESSION_PID=<pid>`, the
app's own process id, so a claim made from the app names it and its liveness follows
the app's process.

## Startup

1. Flags: `--all`, `--pending`, `--config <path>`, `--version`, and at most one argument, a
   prefix or a task id.
2. The config, `$XDG_CONFIG_HOME/tasks-tui/config.toml` (else `.config/` under the
   home directory) or the `--config` path. Absent means defaults.
3. The probe, `tasks projects`. Its rows give every prefix's registered root, used
   for checkouts and to match the current directory.
4. Seed the roots and project-color map from that response, before resolving
   direct Project/Task entry. No familiar files or project git remotes are read.
5. Resolve opening views: a prefix or deepest registered root opens that project;
   `--all` opens Projects; a task id uses the selection below.
6. Bubble Tea reports the terminal profile. TrueColor uses project RGB, ANSI256
   the supplied index, other profiles neutral. Only ANSI256/TrueColor requests
   foreground/background; startup explicitly sets the one-second color-query
   timeout. These requests have no second terminal reader or theme-file access.

Warnings in the probe's and the lookups' responses are dropped: the views they open
make the same calls on their first load and report them.

## Checkout selection

A task's checkout is chosen from the row or response tasks-tui holds for it:

1. `park.worktree`, when the task is parked and that directory exists;
2. else `claim.worktree`, when the claim has `live: true` and that directory exists;
3. else the project's registered root from `projects`.

A recorded worktree that is missing falls through, with the notice
`parked checkout <dir> is gone; using <dir>` (or `claimed checkout …`): in the status
line for a write or launch, and among the warnings of the Task view or task form.
A stale claim never selects a checkout. The Task view starts from its row's park and
claim and, once `show` returns, uses the response's `park` and `claim`; the task form
keeps the checkout it was opened with for both its `show` and its `edit`.

`show`, `edit`, `start`, `park`, `done` and `drop` run with `-C <checkout>`, and a
launch opens there. `projects`, `prime`, `ready`, `list`, `tags` and `add` run without
`-C`. `tasks-tui <id>` needs an id of the form `<prefix>-<6 hex>` whose prefix is
registered. It runs `list --project <prefix> --parked` first, since a park is
readable from anywhere while a record that exists only on its branch is not; when
the id is there, it takes the title, park and claim from that row. Otherwise it runs
`show <id>` with `-C <root>` and takes them from the response.

## Reads

| Call | Issued by | Fields used |
| --- | --- | --- |
| `projects` | startup; Projects load | `projects[]`: `prefix`, `root`, `color`, `color_256`, `reachable`, `counts.{doing,todo,idea,blocked}`, `last_activity` |
| `list --all-projects --parked` | Projects load | the number of `tasks[]` |
| `prime --project <p>` | Projects pane; every Project load | `doing[]` and `ready[]` list rows, `parked[]` parked rows (pane); `counts` (Project header) |
| `ready --project <p>` | Ready tab | `tasks[]` list rows |
| `list --project <p> [--status doing\|idea\|done] [--sort updated]` | Doing, Open, Ideas, Done tabs | `tasks[]` list rows |
| `list --project <p> --parked` | Doing tab; `tasks-tui <id>` | `tasks[]` parked rows |
| `list --all-projects --pending` / `list --project <p> --pending` | Pending queue view | `observed_at`, `coverage.{complete,unavailable_projects}`, `items[]` (`decision` and `legacy_wait` items with nullable `opened_at`/`age_seconds`), `warnings` |
| `show <id>` | Task view; edit form; `tasks-tui <id>`; human-request park | `task.{id,title,status,priority,size,complexity,process,owner,created,started,updated,completed,tags,source,spec,plan,step,body,notes[]}`, `spec_path`, `plan_path`, `step_found`, `parent`, `children[]`, `depends_on[]`, `claim`, `park` (with optional `decision_id`), `escalation`, `periodic.{every,due}`, optional `decisions[]` |
| `tags --project <p>` | task form | `tags[].tag`, for completion |

The task view reads optional `show.actions[]` from the same detail call. It
displays each action's actor and availability and all blocking reason codes with
their identifiers where present. The client retains `signals`, the answer
`subject`, and the curate `window`. Detail remains usable when the installed
tracker omits the array. The Ready tab continues to use the tracker’s ready
selection, and the usual commands handle writes.

Every response's `warnings[]` is shown (see the guide's
[Messages and warnings](guide.md#messages-and-warnings)), except the startup and
form-tag lookups'.

Project rows also decode optional `color` (exactly `#` plus six hexadecimal
digits) and `color_256` (an integer in 16..255, supplied by tasks). Null or absent
color metadata is accepted. RGB without an index is valid; an index without RGB,
an invalid RGB or index, or a wrong JSON type is a typed decode error. The TUI
does not derive a replacement project index. These fields come only from the CLI.

A **list row** gives `id`, `title`, `status`, `priority`, `size`, `complexity`,
`process`, `updated`, `tags`, `open_descendant_count`, `periodic.every`, `claim`
(`owner`, `worktree`, `live`) and `park` (`next_step`, `waiting_on`, `worktree`). A
**parked row** is a separate shape with the same fields used except `periodic`, which
it lacks: only `id`, `title` and `park` are required, because a park record can
outlive the task file in the checkout that answers. A parked row without `status` is
unresolved: it is shown from the park alone and offers no transition or launch.

A claim with `live: false` is shown as stale and never counts as ownership (see
[Known gaps](#known-gaps)).

**Decoding.** A response must be a JSON object. Unknown fields are ignored, and a
value of the wrong JSON type is a decode error. The keys below must be present and
not null, whether or not the app displays them:

| Shape | Required, not null |
| --- | --- |
| `projects` response | `projects`, `warnings` |
| `prime` response | `prefix`, `counts`, `ready`, `doing`, `parked`, `warnings` |
| `ready`, `list` response | `tasks`, `warnings` |
| `show` response | `task`, `warnings` |
| `tags` response | `tags`, `warnings` |
| project row | `prefix`, `root`, `reachable` |
| counts | `idea`, `todo`, `doing`, `blocked`, `shelved`, `done`, `dropped` |
| list row | `id`, `title`, `status`, `priority`, `updated`, `child_count`, `open_descendant_count` |
| parked row | `id`, `title`, `park` |
| task | `id`, `title`, `status`, `priority`, `parallel`, `created`, `updated`, `body` |
| claim | `owner`, `session`, `host`, `worktree`, `started`, `seen`, `live` |
| park | `at`, `next_step`, `waiting_on`, `session`, `owner`, `host`, `worktree` |
| periodic | `every`, `due_now` |
| parent, child | `id`, `title`, `status` |
| dependency | `id`, `resolved` |
| note | `at`, `by`, `text` |
| escalation | `level`, `at`, `session` |
| tag | `tag`, `count`, `projects` |

A project row's `counts`, `total` and `last_activity`, and a tag's `meaning`, must be
present but may be null. Every other field may be omitted or null, including empty
collections: a row's `tags`, a task's `tags`, `depends` and `notes`, and `show`'s
`children` and `depends_on`. A nested object's keys are checked only when the object
is present and not null. An array's elements must be objects: a `null` element is a
decode error.

## Writes

| Action | Command | Response |
| --- | --- | --- |
| add | `add <title> --project <p> [-b <body>] [--tag <t>]…` | `id`, `action`, `warnings` |
| edit | `-C <checkout> edit <id> --title <title> -b <body> --no-tags [--tag <t>]…` | `id`, `warnings` |
| start | `-C <checkout> start [--force] <id>` | `id`, `warnings` |
| park | `-C <checkout> park <id> <next step> [--waiting-on user] [--reason <r>] [--decision <uuid>] [--decision-kind <k> --decision-origin <o> --decision-text <t> [--artifact-repo <r> --artifact-path <p> --artifact-commit <c> \| --artifact-path <p> --artifact-sha256 <s>]]` | `id`, `warnings` |
| done | `-C <checkout> done <id> [<message>]` | `id`, `warnings` |
| drop | `-C <checkout> drop <id> [<message>]` | `id`, `warnings` |
| decision answer | `-C <checkout> decision answer <id> <uuid> --accept\|--decline\|--reply <text> [--seen-scope <token>]` | `id`, `action`, `replayed`, `decision`, `resolution`, `warnings` |
| decision withdraw | `-C <checkout> decision withdraw <id> <uuid> --reason <text>` | `id`, `action`, `replayed`, `decision`, `resolution`, `warnings` |

`--project` is always passed on `add`, so the app's working directory never decides
where a task lands. `add` passes no `--agent`, no `--status` and no other field: a
person files from the app. `edit` always passes the body and replaces the tag set.
`--force` is passed only after a `claimed` refusal and the person's `F`. A success's
warnings are logged and the top view reloads; no write patches the view in place.

## Launch handoff

1. The prompt is the config's `launch.prompt` with `{id}` and `{title}` replaced.
2. The argv is `launch.terminal` with `{dir}` replaced by the checkout, followed by
   the harness's `command` with `{prompt}` replaced by the prompt.
3. The environment is tasks-tui's own, unscrubbed, less `TASKS_SESSION` and
   `TASKS_SESSION_PID`, plus the harness's `env` table; a key in that table overrides
   an inherited one. Agent variables such as `TASKS_AGENT` and `TASKS_MODEL` are
   kept: the harness is an agent.
4. The process starts in the checkout, in its own session (`setsid`), with standard
   input, output and error on the null device. A background goroutine reaps it;
   the app never blocks on it.

Launch never changes the task: the prompt asks the agent to run `tasks start`, so the
claim belongs to the agent's session. The default terminal is
`kitty --directory {dir} --`, and the default harness commands are:

| Harness | Command |
| --- | --- |
| claude | `claude {prompt}` |
| codex | `codex {prompt}` |
| opencode | `opencode --prompt {prompt}` |
| crush | `crush` (no initial prompt) |

An unknown harness, an empty terminal or command, or a failed start is an error;
what the harness does afterwards is not observed.

## Failures

- **The error envelope.** On a non-zero exit, stderr is decoded as
  `{"error": {"kind": "<kind>", "detail": "<detail>"}}`, with `kind` a non-empty
  string and `detail` a string; it is shown as `<kind>: <detail>`.
- **`timeout` / `canceled`.** A call stopped by its context reports the command
  and the context's deadline or cancellation cause.
- **`unparseable`.** A non-zero exit whose stderr is not that envelope (a usage
  error or a panic) has kind `unparseable` and the
  trimmed stderr as detail.
- **`spawn`.** A binary that cannot be started: detail `<bin>: <error>`.
- **`decode`.** Exit 0 with a response that is not an object, lacks a required key,
  holds null in one, or has a mistyped value: detail `response missing <keys>; null
  <keys>` or the JSON error.
- **`claimed`** on `start` offers `start --force`. **`stale_scope`** on `decision answer`
  logs a warning and reloads the view so the person sees the updated scope before making a
  fresh explicit answer; it never retries automatically. Every other refusal (a `done` with
  open descendants, for one) is shown as is; nothing is retried, defaulted or forced.
- **Where they show.** A failed load is an error in the status line and the view
  keeps its previous data; a failed write reads `<action> <id>: <kind>: <detail>`; a
  failed form load or save is shown inside the form.
- **Fatal at startup**, printed as `tasks-tui: <message>` to stderr with exit 1
  before the screen opens: a bad flag or more than one argument; a config that is
  malformed, has an unknown key, or has an empty `tasks`, negative `refresh_seconds`,
  empty `launch.terminal` or empty harness command;
  `projects` failing for any reason above; an argument that is
  neither a registered prefix nor a task id, or whose prefix is unregistered; and a
  failed lookup for `tasks-tui <id>`. Nothing is logged to a file.

## Known gaps

- **No version floor.** tasks-tui neither reads nor requires a `tasks` version, so
  this contract is unversioned; drift shows up only as `decode` errors.
- **Liveness is one boolean.** Claims come back live or not, and `live` is the only
  signal; a stale claim is still returned.
- **Unresolved parked rows offer nothing.** The row carries no `status`, so the app
  offers nothing on it, even though the park's worktree might resolve it. A task
  opened by id from such a row is not marked unresolved: its `show` and transitions
  run in the park's worktree and succeed or fail there.
- **No file watching.** Changes made elsewhere appear only when the view reloads: on
  the refresh timer, `F5`, a write, or opening or returning to the view. With
  `refresh_seconds = 0` there is no timer.
- **A harness without `{prompt}` starts without one.** The prompt is copied to the
  clipboard instead, for the person to paste.
