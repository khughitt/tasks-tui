---
id: tui-0f5acb
title: Rewrite t@t commit authorship in this repository
status: done
priority: 2
size: s
complexity: low
process: direct
owner: main
created: 2026-10-16T16:33:15Z
updated: 2026-10-17T01:04:31Z
started: 2026-10-17T01:03:20Z
completed: 2026-10-17T01:04:31Z
depends: [ops-e5b39c]
tags: []
source: git-identity-leak-2026-10-07
agent: claude-code/claude-opus-5-5
---

Rewrite this repository's t <t@t> commits (from the 2026-10-07 global-config leak; see ops-5749d9) to Keith Hughitt <keith.hughitt@gmail.com>, following the runbook ops-e5b39c. Inventory on titan 2026-10-16: 16 t@t commits on all refs, 7 unpushed on main, 1 worktree(s); remote: khughitt/tasks-tui.

## Notes

- 2026-10-17T01:03:20Z (main): started
  provenance: {"harness_session":"claude-code:c6863573-34c2-470d-b394-259f7cc75503","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-10-17T01:03:21Z (main): Direct history maintenance under ops-e5b39c's pilot procedure (as tasks-bce5ec). No agent session in tui on Titan or Europa; one worktree (main) on both hosts. Inventory: 161 commits, 16 t@t, all on main (origin/main reaches 9, pushed). No tags, stash, signatures, extra headers, repo-local user.*. Remote khughitt/tasks-tui main 3df4bfe100f80d099cc8efdc3e19f16c8902b7a4 (lease). Steps: bundle --all; mirror; filter-repo --mailmap; byte-verify; citation check in every project; guarded update-ref; attach map; test-fast; push main --force-with-lease.
- 2026-10-17T01:03:41Z (main): attached: 2026-10-16-identity-commit-map.txt (1640 bytes): Authorship rewrite map: 20 changed commit IDs (old new); resolves historical SHA references in task notes and docs without editing them
- 2026-10-17T01:04:31Z (main): done
  provenance: {"harness_session":"claude-code:c6863573-34c2-470d-b394-259f7cc75503","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-10-17T01:04:31Z (main): Rewrote 16 t@t commits to Keith Hughitt (32 identity lines, 20 hashes changed) in one filter-repo pass in a mirror outside Dropbox; verified all 163 commits byte-for-byte (only mapped parents and identities differ; every tree preserved). Swapped main in one guarded update-ref transaction. Rewritten SHAs cited anywhere: this task's lease note and an ops push receipt (provenance); map attached (tasks/files/tui-0f5acb/2026-10-16-identity-commit-map.txt). just test-fast green. Pushed main with lease 3df4bfe -> 5ed8e6b (khughitt/tasks-tui). A fresh remote mirror, Titan and Europa show 0 t@t; Europa fsck clean. Forced only past ops-e5b39c, whose pilot procedure was followed. Artifacts on Titan at WORK_ROOT/tasks-tui/identity-rewrite-20261017T010322Z.
  provenance: {"harness_session":"claude-code:c6863573-34c2-470d-b394-259f7cc75503","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
