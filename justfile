# Front door for tests. Gates: `just check` (seconds) at pre-commit, `just gate` (check +
# suite) at pre-push. The hooks in .githooks/ call `hook-pre-commit` and `hook-pre-push`,
# which run the same commands under their own target names for the cross-project audit.
# Every recipe runs through the vendored timing wrapper tools/tt (source: ops bin/tt).

set quiet

tt := "python3 tools/tt"

fast_cmd := "go test $(go list ./internal/... | grep -v '/integration$')"
test_cmd := "go test ./..."
check_cmd := "python3 tools/ops-check && test -z \"$(gofmt -l .)\" && go vet ./... && go tool staticcheck ./... && tasks check"

# Dependencies and hooks; run once after `git worktree add`.
setup:
    go mod download
    git config core.hooksPath .githooks

# The packages without the integration test.
test-fast:
    {{tt}} test-fast -- sh -c '{{fast_cmd}}'

test:
    {{tt}} test -- sh -c '{{test_cmd}}'

smoke:
    tools/smoke-tui

# Recapture one tasksctl fixture from live output, scrubbed of this machine's layout by
# tools/scrub-fixtures (registry, hostname, WORK_ROOT); raw tasks output never lands in
# testdata. The tracked fixtures came from: projects `projects`; prime `prime --project
# tasks`; list_parked `list --all-projects --parked`; show `show <id>`. list_rows.json and
# error_claimed.json are hand-shaped (claim liveness, a claimed error) and are edited, not
# recaptured. Example: just capture show show tui-ce6c9f
capture name +args:
    #!/bin/sh
    set -eu
    raw=$(mktemp) && trap 'rm -f "$raw"' EXIT
    tasks {{args}} > "$raw"
    tools/scrub-fixtures < "$raw" > internal/tasksctl/testdata/{{name}}.json.tmp
    mv internal/tasksctl/testdata/{{name}}.json.tmp internal/tasksctl/testdata/{{name}}.json

# Seconds: hygiene, format, vet, staticcheck, tasks check.
check:
    {{tt}} check -- sh -c '{{check_cmd}}'

gate: check test

hook-pre-commit:
    {{tt}} hook-pre-commit -- sh -c '{{check_cmd}}'

hook-pre-push:
    {{tt}} hook-pre-push -- sh -c '{{check_cmd}} && {{test_cmd}}'

install:
    #!/bin/sh
    set -eu
    go install ./cmd/tasks-tui
    tui_bin_dir="$(go env GOBIN)"
    if [ -z "$tui_bin_dir" ]; then tui_bin_dir="$(go env GOPATH)/bin"; fi
    ln -sf tasks-tui "$tui_bin_dir/tui"
