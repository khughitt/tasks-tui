# Task-list sorting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add local, keyboard-selected sorting to Project-view task tabs.

**Architecture:** Keep sort and selector state on `projectView`. `cycleSort(key)` owns the three-state local order; the selector only chooses its key. Extend the existing table header/measurement path with one decorated-label floor, and cache render-time task-table widths for selector input. No `tasks` command, config, or Projects-view behavior changes.

**Tech Stack:** Go, Bubble Tea v2, Lip Gloss v2; no dependencies.

**Spec:** `docs/specs/2026-09-15-task-list-sorting-design.md`

## Global Constraints

- Sort only the visible Project-view data columns; gutter and marks are structural.
- Preserve tracker order until a local sort is selected; missing values remain last in either direction.
- `S` captures keys only in a loaded Project view; selector capture and filtering are mutually exclusive.
- Run `just check` before each commit and `just gate` before task completion.

---

### Task 1: Local sort model and measured headers

**Files:**

- Modify: `internal/ui/project.go`, `internal/ui/columns.go`, `internal/ui/rows.go`
- Test: `internal/ui/project_test.go`, `internal/ui/columns_test.go`

**Interfaces:**

- `func (v *projectView) cycleSort(key string)` changes `key` through ascending, descending, and off.
- `table.header` receives decorated labels, and `table.measure` receives those labels as width floors.

- [ ] **Step 1: Write failing sort tests**

Add literal rows to `project_test.go` and call `pv.cycleSort("prio")`, asserting `1, 2, 3`, then `3, 2, 1`, then the loaded order. Add cases for `xs<s<m<l<xl`, `low<mid<high`, case-insensitive titles, ISO `Updated`, and missing/unparseable values last. Assert the selected ID survives sorting and filtering.

- [ ] **Step 2: Run the focused tests red**

Run: `go test ./internal/ui -run 'TestProjectCycleSort|TestTaskHeaderDecoration' -count=1`

Expected: FAIL because `cycleSort` and decorated header measurement do not exist.

- [ ] **Step 3: Implement the minimum local model**

In `project.go`, keep `sortKey string` and `descending bool`; implement `cycleSort`. Make `applyFilter` rebuild matching rows then stable-sort only when `sortKey != ""`; compare ranked fields by fixed maps (comment that status rank follows the tracker enum), `age` by parsed RFC3339, case-fold titles, and use ID as the final comparison. `layoutRows` computes decorated labels once and passes them to both `measure` as per-column floors and `header`; do not add a callback or duplicate table drop logic.

- [ ] **Step 4: Run focused tests green**

Run: `go test ./internal/ui -run 'TestProjectCycleSort|TestTaskHeaderDecoration' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit and complete the model child**

```bash
tasks done tui-c59419 "Added local task-list sorting and measured headers"
git add internal/ui/project.go internal/ui/columns.go internal/ui/rows.go internal/ui/project_test.go internal/ui/columns_test.go
git commit -m "feat: sort task lists locally"
```

### Task 2: Selector, capture, and legend

**Files:**

- Modify: `internal/ui/project.go`, `internal/ui/columns.go`, `internal/ui/rows.go`, `internal/ui/keys.go`
- Test: `internal/ui/project_test.go`, `internal/ui/app_test.go`
- Modify: `docs/specs/2026-09-14-tasks-tui-v1.1-polish-design.md`, `docs/specs/2026-09-15-task-list-sorting-design.md`

**Interfaces:**

- `projectView.capturing()` returns true for filtering or an open selector; `render` caches `layoutRows(...).widths` on the view for `update`.
- `S`, left/right, Enter, Esc, and backspace follow the approved selector contract.

- [ ] **Step 1: Write failing interaction tests**

Drive a loaded Project view: press `S`, assert the first/active header is selected; left/right skip structural columns; Enter calls `cycleSort` and closes. Assert Esc and backspace close without popping the view, non-selector keys do nothing, `S` remains filter text, and an active sort re-applies after reload and tab switch. Feed a narrow `tea.WindowSizeMsg`, render once, then press `S` to prove dropped columns are skipped.

- [ ] **Step 2: Run interaction tests red**

Run: `go test ./internal/ui -run 'TestProject.*Selector|TestCapturing' -count=1`

Expected: FAIL because `S` is only handled by the global Projects sort binding.

- [ ] **Step 3: Implement selector routing and rendering**

Handle `S` in `projectView.update` before normal movement when data is loaded. While selecting, consume left/right, Enter, Esc, and backspace; leave every other key unchanged. In `render`, cache the current `layoutRows(...).widths`; in `update`, enumerate only nonzero-width data columns from that cache. Render candidate headers with `v.env.Styles.Surface(v.env.slot(v.prefix))`. Change `keys.Sort` help to `sort` and update the superseded v1.1 legend text. Keep `projectsView`'s existing `S` behavior.

- [ ] **Step 4: Run interaction tests green**

Run: `go test ./internal/ui -run 'TestProject.*Selector|TestCapturing|TestEveryGroupedBindingHasHelp' -count=1`

Expected: PASS.

- [ ] **Step 5: Verify and complete**

Run: `just gate`

Then update the sorting spec status to `implemented`, run `tasks done tui-cfa2bf "Added the task-list selector and capture"`, and commit the task record and documentation with the code. Once `tasks prime` offers `tui-62bb3f` under closeout, confirm the goal and run `tasks done tui-62bb3f "Added local keyboard sorting for Project task lists"`.
