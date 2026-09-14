# tasks-tui

A keyboard-driven terminal front end for the `tasks` tracker, on Bubble Tea. It shells
out to the `tasks` binary and reads its JSON; it never touches `tasks/*.md` itself.

    just setup      # deps and git hooks
    just install    # go install ./cmd/tasks-tui
    tasks-tui       # Projects view, or the project you are standing in
    tasks-tui tui   # one project by prefix
    tasks-tui tui-d6e352

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
