# Front door for tests. Focused: `just test-one <runner args>`. Fast: `just test-fast`.
# Full suite: `just test`. Gates: checks at pre-commit, check + suite at pre-push.
# Hooks time mandatory hygiene once; broad children record their own policy decisions.
# Leaf recipes run through tools/tt (source: ops bin/tt); compositions are untimed.

set quiet
set positional-arguments

tt := "python3 tools/tt"

fast_cmd := "go test $(go list ./internal/... | grep -v '/integration$')"
test_cmd := "go test -count=1 -v ./..."
one_cmd := "go test"
check_hygiene_cmd := "python3 tools/ops-check && python3 tools/ops-docs check && tasks check"
ci_hygiene_cmd := "python3 tools/ops-check && python3 tools/ops-docs check"
check_code_cmd := "test -z \"$(gofmt -l .)\" && go vet ./... && go tool staticcheck ./..."

# README and design documents stay on the full check; only the agent guide skips Go checks.
docs_paths := "AGENTS.md"
docs_check_cmd := check_hygiene_cmd

# Remote CI suite declaration for .githooks/pre-push. Keep ci_suite_refs empty
# until remote CI runs the broad suite for the matching refs.
ci_remote := "origin"
ci_suite_refs := ""

# Dependencies and hooks; run once after `git worktree add`.
setup:
    {{tt}} setup -- sh -c 'go mod download && git config core.hooksPath .githooks'

# The packages without the integration test (policy-controlled).
test-fast:
    {{tt}} test-fast -- sh -c '{{fast_cmd}}'

# Focused package or test pattern forwarded to `go test` (always runs).
test-one +args:
    {{tt}} test-one -- {{one_cmd}} "$@"

# Full test suite, including the real-binary integration test (policy-controlled).
test:
    {{tt}} test -- sh -c '{{test_cmd}} 2>&1'

# Deliberate full-suite verification; only the inner test target records timing.
test-full:
    TT_FORCE=1 just test

smoke:
    tools/smoke-tui

# Regenerate the identity and family regions of README.md and AGENTS.md from
# identity.toml and tools/family.toml (ops docs/project-guides.md).
docs:
    {{tt}} docs -- python3 tools/ops-docs write

# Recapture one tasksctl fixture from live output, scrubbed of this machine's layout by
# tools/scrub-fixtures (registry, hostname, WORK_ROOT); raw tasks output never lands in
# testdata, and the values kept verbatim are listed on stderr for review. The tracked
# fixtures came from: projects `projects`; prime `prime --project
# tasks`; list_parked `list --all-projects --parked`; show `show <id>`. list_rows.json and
# error_claimed.json are hand-shaped (claim liveness, a claimed error) and are edited, not
# recaptured. Example: just capture show show tui-ce6c9f
capture name +args:
    #!/bin/sh
    set -eu
    raw=$(mktemp) && trap 'rm -f "$raw"' EXIT
    tasks {{args}} > "$raw"
    tools/scrub-fixtures --report < "$raw" > internal/tasksctl/testdata/{{name}}.json.tmp
    mv internal/tasksctl/testdata/{{name}}.json.tmp internal/tasksctl/testdata/{{name}}.json

# Mandatory local hygiene, including detached task-record consistency.
check-hygiene:
    {{tt}} check-hygiene -- sh -c '{{check_hygiene_cmd}}'

# Hosted runners have no private detached task records.
ci-check-hygiene:
    {{tt}} ci-check-hygiene -- sh -c '{{ci_hygiene_cmd}}'

# Broad Go formatting and static analysis checks (policy-controlled).
check-code:
    {{tt}} check-code -- sh -c '{{check_code_cmd}}'

# Hygiene + Go static analysis.
check:
    just check-hygiene
    just check-code

gate: check test

# Hook timers cover mandatory hygiene only; broad children record their own decisions.
# Split-hook durations are incomparable with the earlier whole-hook series.
hook-pre-commit:
    {{tt}} hook-pre-commit -- sh -c '{{check_hygiene_cmd}}'
    just check-code

# Guide-only changes retain hygiene and task consistency checks.
hook-pre-commit-docs:
    {{tt}} hook-pre-commit-docs -- sh -c '{{docs_check_cmd}}'

# Uncovered pushes retain the same children as gate.
hook-pre-push:
    {{tt}} hook-pre-push -- sh -c '{{check_hygiene_cmd}}'
    just check-code
    just test

# CI carries the full suite only when every pushed ref is covered on ci_remote.
hook-pre-push-fast:
    {{tt}} hook-pre-push-fast -- sh -c '{{check_hygiene_cmd}}'
    just check-code
    just test-fast

install:
    #!/bin/sh
    set -eu
    go install ./cmd/tasks-tui
    tui_bin_dir="$(go env GOBIN)"
    if [ -z "$tui_bin_dir" ]; then tui_bin_dir="$(go env GOPATH)/bin"; fi
    ln -sf tasks-tui "$tui_bin_dir/tui"
