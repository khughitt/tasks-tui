---
id: tui-a128d7
title: Tag completion in quick add from tasks tags
status: done
priority: 2
size: m
complexity: mid
process: direct
owner: feat/tag-completion
created: 2026-09-14T12:42:09Z
updated: 2026-09-14T13:26:22Z
started: 2026-09-14T13:14:12Z
completed: 2026-09-14T13:26:22Z
depends: []
tags: [v2]
agent: codex
---

Why: reuse the project vocabulary while filing tasks without leaving quick add. Done: asynchronously fetch tasks tags --project <prefix> through the typed client; offer matching tags while typing a trailing #prefix before the body separator; Tab accepts and Up/Down choose using the existing textinput suggestions. Suggestions follow the default or explicit >project, exclude tags already entered, preserve Unicode and all surrounding input, and allow arbitrary new tags. Late lookup results after project changes or prompt close/reopen cannot change suggestions or log stale failures; current errors/warnings enter the session log. Empty tag lists remain usable. Completion is at the end of input, not a new mid-line editor. Tests cover JSON boundary/null handling, token/body/project context, actual key/paste acceptance and saved argv, and stale request behavior; one isolated real-terminal smoke verifies the installed command. Where: internal/ui/quickadd.go, overlay.go, app.go; internal/tasksctl/client.go and decode.go; internal/quickadd/parse.go; installed textinput suggestion API. Direct process: existing tracker contract and native suggestion behavior settle the bounded implementation, with focused tests as the acceptance check. Original: deferred from v1 (docs/specs/2026-09-13-tasks-tui-v1-design.md §14).

## Notes

- 2026-09-14T13:14:12Z (feat/tag-completion): scope: scoped; promoted to direct implementation using native textinput suggestions over project-scoped tasks tags, with asynchronous lookup and explicit stale-result checks; worktree .worktrees/tag-completion.
- 2026-09-14T13:22:37Z (feat/tag-completion): Tracker tag names are not restricted to quick-add syntax. Completion filters with the existing tag grammar so accepting a suggestion cannot insert whitespace or flags into the title; a regression failed before the guard and now passes.
- 2026-09-14T13:26:22Z (feat/tag-completion): Quick add now completes trailing tags with Tab and Up/Down from project-scoped tasks tags, preserves manual entry, and rejects stale lookup results. Typed decoder, grammar, UI lifecycle tests and installed tui PTY smoke passed; independent review clean.
