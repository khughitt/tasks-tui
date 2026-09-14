# tasks-tui

A keyboard-driven terminal front end for the `tasks` tracker, on Bubble Tea. It shells
out to the `tasks` binary and reads its JSON; it never touches `tasks/*.md` itself.

    just setup      # deps and git hooks
    just install    # installs tasks-tui and the short command tui
    tui             # Projects view, or the project you are standing in
    tui tui         # one project by prefix
    tui tui-d6e352

Both commands are installed in Go's binary directory, which must be on `PATH`.

Keys: `?` legend · `a` quick add · `l` launch an agent · `s` start · `p` park · `d` done · `x` drop · `W` messages · `q` quit.

Quick add: `#tag !2 ~m ^mid @30d >prefix · ? first = idea · -- body`.

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
