---
id: tui-bbccd2
title: "staticcheck reads Go 1.27.2 export data: raise golang.org/x/tools to v0.51.0"
status: done
priority: 1
size: xs
complexity: low
process: direct
owner: fix/staticcheck-export-v5
created: 2026-10-09T14:35:15Z
updated: 2026-10-09T14:37:19Z
started: 2026-10-09T14:35:15Z
completed: 2026-10-09T14:37:19Z
depends: []
tags: []
source: "https://github.com/dominikh/go-tools/issues/1832"
model: claude-opus-5-5
agent: claude-code/claude-opus-5-5
---

pacman upgraded go 1.27.1 -> 1.27.2 on 2026-10-09; its compiler writes export data V5 and staticcheck v0.8.1's pinned x/tools (v0.44.1-0.20260420230617) reads at most V4, so `go tool staticcheck` fails in check and the pre-push hook. Upstream: dominikh/go-tools#1832, open; v0.8.1 is the newest release and master fails the same way. x/tools v0.51.0 adds pkgbits V5. Local workaround: require x/tools v0.51.0; retire it (drop the explicit bump) once a staticcheck release depends on an x/tools with V5.

## Notes

- 2026-10-09T14:35:15Z (main): started
  provenance: {"harness_session":"claude-code:4622fdd1-aa60-41b9-b21d-d95d5983a002","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-10-09T14:36:34Z (fix/staticcheck-export-v5): resumed
  provenance: {"harness_session":"claude-code:4622fdd1-aa60-41b9-b21d-d95d5983a002","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-10-09T14:37:19Z (fix/staticcheck-export-v5): done
  provenance: {"harness_session":"claude-code:4622fdd1-aa60-41b9-b21d-d95d5983a002","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
- 2026-10-09T14:37:19Z (fix/staticcheck-export-v5): x/tools v0.51.0 (pkgbits V5) lets staticcheck v0.8.1 read Go 1.27.2 export data; check and test-fast pass. Retire the explicit bump when a staticcheck release carries it (dominikh/go-tools#1832).
  provenance: {"harness_session":"claude-code:4622fdd1-aa60-41b9-b21d-d95d5983a002","harness_session_source":"CLAUDE_CODE_SESSION_ID"}
