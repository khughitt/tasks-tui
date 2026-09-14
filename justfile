# Front door for tests. Gates: `just check` (seconds) at pre-commit, `just gate` (check +
# suite) at pre-push. The hooks in .githooks/ call `hook-pre-commit` and `hook-pre-push`,
# which run the same commands under their own target names for the cross-project audit.
# Every recipe runs through the vendored timing wrapper tools/tt (source: ops bin/tt).

tt := "python3 tools/tt"

fast_cmd := "go test ./internal/..."
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

# Seconds: hygiene, format, vet, staticcheck, tasks check.
check:
    {{tt}} check -- sh -c '{{check_cmd}}'

gate: check test

hook-pre-commit:
    {{tt}} hook-pre-commit -- sh -c '{{check_cmd}}'

hook-pre-push:
    {{tt}} hook-pre-push -- sh -c '{{check_cmd}} && {{test_cmd}}'

install:
    go install ./cmd/tasks-tui
