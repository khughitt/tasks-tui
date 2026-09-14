---
id: tui-8b852e
title: Reject null park data in parked-row decoding
status: done
priority: 1
size: xs
complexity: low
process: direct
owner: design/v1
created: 2026-09-14T12:48:50Z
updated: 2026-09-14T12:51:10Z
started: 2026-09-14T12:49:14Z
completed: 2026-09-14T12:51:10Z
depends: []
parent: tui-d6e352
tags: [v1]
agent: codex
---

Final branch review found ParkedRow accepts park:null but the parked pane requires Park.NextStep. Require non-null park at the JSON boundary, add a decoder regression, and correct the matching plan example.

## Notes

- 2026-09-14T12:51:10Z (design/v1): reject parked rows with null park data at decode boundary and align the implementation plan
