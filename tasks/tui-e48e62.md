---
id: tui-e48e62
title: "README: a screenshot or short recording of the Projects and task views"
status: todo
priority: 3
size: s
complexity: mid
process: direct
created: 2026-09-19T13:13:54Z
updated: 2026-10-17T20:58:53Z
depends: []
tags: []
agent: "claude-code/claude-opus-5[1m]"
---

Why: the public README (github.com/khughitt/tasks-tui) is text only; one image of the Projects view and one of a task view with the launch picker (`c c`) would show what the tool is faster than the key table.

Done:
- A script under `tools/` (a `just` recipe beside `smoke`) regenerates the images reproducibly: it builds a throwaway registry under a temp `XDG_CONFIG_HOME`/`XDG_STATE_HOME` (as `tools/smoke-tui` does) holding 3–4 invented demo projects with plausible tasks across statuses, runs `tasks-tui --all` in a private tmux socket at a fixed size, drives it with `send-keys`, captures with `capture-pane -e`, renders the ANSI to images, and reaps tmux on EXIT/INT/TERM.
- Two images committed (e.g. `docs/images/`) and referenced near the top of README.md; no real project prefix, path, hostname, or task text appears in them.
- The user has looked at the rendered images and accepted them before the commit.

Where: `tools/smoke-tui` (headless tmux pattern), README.md, go.mod's `tool` directives (staticcheck is one). Renderer lean: charmbracelet `freeze` as a `go tool` (ANSI → SVG/PNG, no ttyd/ffmpeg); a recording (vhs) is out of scope unless the stills fall short. None of freeze, vhs, or asciinema is installed on this host.
