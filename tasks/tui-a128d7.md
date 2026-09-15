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
updated: 2026-09-15T21:20:41Z
started: 2026-09-15T21:18:03Z
completed: 2026-09-15T21:20:41Z
depends: []
tags: [v2]
agent: codex
---

Why: complete quick-add tags from the tracker without leaving the prompt. Done: fetch project-scoped tags through the typed client; use native text-input suggestions for a trailing #prefix, preserving arbitrary manual tags and discarding stale lookups. Where: internal/quickadd, internal/tasksctl, and internal/ui/quickadd. Check: focused client, grammar, and UI lifecycle tests plus just gate. Existing evidence: the complete, tested implementation is commit 174855b on feat/tag-completion; adapt it to current main in its existing worktree.

## Notes

- 2026-09-15T21:17:26Z (main): scope: scoped; promoted to direct implementation from the established 174855b branch, retaining its typed boundary, native suggestions, and stale-result checks
- 2026-09-15T21:20:41Z (feat/tag-completion): Complete project-scoped trailing-tag suggestions in quick add via tasks tags, with typed decoding and stale-request protection
