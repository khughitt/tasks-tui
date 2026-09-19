# tasks-tui

A keyboard-driven terminal front end for the [tasks](https://github.com/khughitt/tasks)
tracker, on Bubble Tea. It shells out to the `tasks` binary and reads its JSON; it never
touches `tasks/*.md` itself, so the CLI stays the only writer and the whole contract.

## Install

Needs Go and a `tasks` binary on `PATH` (install it from the tasks repository first).

    just setup      # deps and git hooks
    just install    # installs tasks-tui and the short command tui

Both commands land in Go's binary directory (`go env GOBIN`, else `$GOPATH/bin`), which
must be on `PATH`.

## Use

    tui                 # Projects view, or the project you are standing in
    tui --all           # Projects view even inside a registered project
    tui tui             # one project by prefix
    tui tui-d6e352      # one task
    tui --config path   # a config file other than the default

A project opens on five tabs: Ready, Doing, Open, Ideas, Done.

## Keys

`?` shows the legend in the app. Chords (`c …`, `s …`, `S …`, `g g`) wait up to 900 ms
for their second key.

| | |
|---|---|
| move | `j`/`k` or Up/Down · `h`/`l` or Left/Right switch tabs · `tab`/`shift+tab` cycle tabs · `1`–`5` pick a tab · `pgup`/`pgdn` page · `g g`/Home top · `G`/End bottom |
| open | `enter`/`i` open · `esc` back · `/` filter · `F5` reload |
| sort | `s p` priority · `s a` age · `s t` title, ascending; `S` as the prefix sorts descending. In Projects, `p` is prefix and `a` is activity |
| task | `a` add · `e` edit · `space` start · `p` park · `d` done · `x` drop · `y` copy id |
| launch | `c l` Claude · `c o` Codex · `c r` Crush · `c O` OpenCode · `c c` picker |
| app | `?` legend · `W` console · `q` quit (or close a panel) |

Add and edit open a centered form for title, body, and tags. Tab or Shift+Tab move
between fields, Ctrl+S saves, Escape cancels. In tags, Enter accepts a suggestion and
Up/Down choose among matches from the current project.

Launching opens a new terminal in the task's checkout and hands the harness a prompt
naming the task; the prompt and every command are configurable below.

## Config

`~/.config/tasks-tui/config.toml` (every key optional; unknown keys are an error):

```toml
tasks = "tasks"
refresh_seconds = 30

[launch]
terminal = ["kitty", "--directory", "{dir}", "--"]
prompt = "Run `tasks start {id}` and continue that task: {title}"

[launch.harness.claude]
command = ["claude", "{prompt}"]

[launch.harness.codex]
command = ["codex", "{prompt}"]

[launch.harness.opencode]
command = ["opencode", "--prompt", "{prompt}"]

[launch.harness.crush]
command = ["crush"]

[identity]
slot = { tui = 2 }
```

`{dir}` is the task's checkout, `{id}` and `{title}` the task, `{prompt}` the rendered
prompt. Each project gets one of twelve colour slots, derived from its git remote (or
root) the way familiar's identity scheme does, so a project looks the same here as in
familiar's pets; `identity.slot` pins a project's slot by prefix, and familiar's own pins
in `~/.config/familiar/identities.yaml` are honoured when that file exists.

## Development

    just check      # hygiene, gofmt, go vet, staticcheck, tasks check (seconds)
    just test       # go test ./...; the integration test skips without tasks on PATH
    just gate       # both

Design: `docs/specs/`. Plans: `docs/plans/`. Agent guide: `AGENTS.md`.
