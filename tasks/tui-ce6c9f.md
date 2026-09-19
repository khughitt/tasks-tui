---
id: tui-ce6c9f
title: Rename the Go module to github.com/khughitt/tasks-tui so go install @latest works
status: todo
priority: 2
size: s
complexity: low
process: direct
created: 2026-09-19T13:13:54Z
updated: 2026-09-19T13:13:54Z
depends: []
tags: []
agent: "claude-code/claude-opus-5[1m]"
---

go.mod declares module tasks-tui, so the public repo cannot be installed with go install github.com/khughitt/tasks-tui/cmd/tasks-tui@latest; only a checkout plus just install works. Rename the module path, rewrite the internal imports, keep the tui symlink recipe, and add the go install line to the README Install section.
