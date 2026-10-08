---
id: tui-d2c96d
title: Apply owner-approved MIT licensing
status: todo
priority: 2
size: s
complexity: low
process: direct
created: 2026-09-19T13:13:54Z
updated: 2026-10-08T11:06:45Z
depends: []
tags: []
agent: "claude-code/claude-opus-5[1m]"
---

Why: make the owner-selected MIT license explicit, matching tasks. Owner decision on 2026-10-08 selected MIT for both repositories; tasks-8e189a records the decision and coordination.

Outcome and approach: add the standard MIT text as root LICENSE with verified copyright holder and year, and add a concise README link. This Go module has no package license field equivalent to Cargo package.license; leave go.mod unchanged. Inspect and preserve third-party notices. Inspection found no LICENSE/COPYING/NOTICE file, and go.mod identifies the module as github.com/khughitt/tasks-tui.

Done: root LICENSE contains the standard MIT text and verified notice, README agrees, and applicable repository gates pass. Use the required .worktrees/ workspace and just setup; run tasks start and trial-arm before implementation. Standard text: https://choosealicense.com/licenses/mit/. Verify notice against repository history; the matching tasks notice uses 2026 Keith Hughitt.

Scope: local licensing and documentation only. No push, release or publication. This separately scoped companion follows tasks-8e189a; implementation is not part of the tracker task.

## Notes

- 2026-10-08T11:06:44Z (main): scope: scoped; owner chose MIT for both repositories on 2026-10-08; P2/s/low/direct companion handoff from tasks-8e189a, reusing the existing license idea; inspected Go module and absence of local license files; implementation remains with the companion project.
