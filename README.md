# tasks-tui

A keyboard-driven terminal front end for the `tasks` tracker, on Bubble Tea. It shells
out to the `tasks` binary and reads its JSON; it never touches `tasks/*.md` itself.

    just setup      # deps and git hooks
    just install    # installs tasks-tui and the short command tui
    tui             # Projects view, or the project you are standing in
    tui tui         # one project by prefix
    tui tui-d6e352

Both commands are installed in Go's binary directory, which must be on `PATH`.

Keys: `?` legend · `enter`/`i` open · `a` quick add · `c c` launch picker · `space` start · `p` park · `d` done · `x` drop · `W` messages · `q` quit (or close a panel).

Launch directly: `c l` Claude, `c o` Codex, `c r` Crush, `c O` OpenCode.
`F5` reloads; `h`/`l` or Left/Right switch tabs; `g g`/Home goes to the top,
`G`/End to the bottom. Sort ascending with `s p` (priority), `s a` (age), or
`s t` (title); use `S` as the prefix for descending. In Projects, `p` sorts
prefix and `a` sorts activity. A pending prefix appears for up to 900 ms.

Quick add: `#tag !2 ~m ^mid @30d >prefix · ? first = idea · -- body`.

When typing a tag at the end of the input, Tab accepts a suggestion and Up/Down
choose among matches from the current project's open tasks. An explicit `>prefix`
changes the project. New tags can always be typed directly; text after `--` is
never completed.

Config: `~/.config/tasks-tui/config.toml` (all optional):

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

Design: `docs/specs/`. Plans: `docs/plans/`.
