package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type markerEvent struct {
	Name        string   `json:"name"`
	Args        []string `json:"args"`
	Stdin       string   `json:"stdin"`
	GitDir      string   `json:"git_dir"`
	GitWorkTree string   `json:"git_work_tree"`
	Literal     string   `json:"literal_pathspecs"`
}

type ttRecord struct {
	Target          string  `json:"target"`
	Execution       string  `json:"execution"`
	ExecutionReason string  `json:"execution_reason"`
	Exit            *int    `json:"exit"`
	Tests           *int    `json:"tests"`
	Selection       *string `json:"selection"`
	Slowest         any     `json:"slowest"`
}

type frontDoorFixture struct {
	t          *testing.T
	root       string
	repo       string
	home       string
	bin        string
	markerFile string
	ttLog      string
}

func checkoutRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func newFrontDoorFixture(t *testing.T) *frontDoorFixture {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	f := &frontDoorFixture{
		t:          t,
		root:       checkoutRoot(t),
		repo:       filepath.Join(base, "repo"),
		home:       filepath.Join(base, "home"),
		bin:        filepath.Join(base, "bin"),
		markerFile: filepath.Join(base, "markers.jsonl"),
		ttLog:      filepath.Join(base, "runs.jsonl"),
	}
	for _, dir := range []string{f.repo, f.home, f.bin} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	f.git("init", "-q")
	f.assertRepoContainment()
	f.writeStubs()
	f.copyRepoFiles("")
	return f
}

func (f *frontDoorFixture) env(extra map[string]string) []string {
	sysPath := os.Getenv("PATH")
	envMap := map[string]string{
		"PATH":                f.bin + string(os.PathListSeparator) + sysPath,
		"HOME":                f.home,
		"XDG_CONFIG_HOME":     filepath.Join(f.home, ".config"),
		"XDG_DATA_HOME":       filepath.Join(f.home, ".local", "share"),
		"XDG_CACHE_HOME":      filepath.Join(f.home, ".cache"),
		"XDG_STATE_HOME":      filepath.Join(f.home, ".local", "state"),
		"TT_LOG":              f.ttLog,
		"MARKER_FILE":         f.markerFile,
		"GIT_CONFIG_GLOBAL":   "/dev/null",
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_AUTHOR_NAME":     "Test User",
		"GIT_AUTHOR_EMAIL":    "test@example.invalid",
		"GIT_COMMITTER_NAME":  "Test User",
		"GIT_COMMITTER_EMAIL": "test@example.invalid",
		"USER":                "testuser",
		"TERM":                "dumb",
		"LANG":                "C.UTF-8",
	}
	for k, v := range extra {
		if v == "" && (k == "TT_FORCE" || k == "GITHUB_ACTIONS") {
			delete(envMap, k)
			continue
		}
		envMap[k] = v
	}
	out := make([]string, 0, len(envMap))
	for k, v := range envMap {
		out = append(out, k+"="+v)
	}
	return out
}

func (f *frontDoorFixture) assertRepoContainment() {
	f.t.Helper()
	top := strings.TrimSpace(f.git("rev-parse", "--show-toplevel"))
	gitDir := strings.TrimSpace(f.git("rev-parse", "--absolute-git-dir"))
	realRepo, err := filepath.EvalSymlinks(f.repo)
	if err != nil {
		f.t.Fatalf("eval repo symlinks: %v", err)
	}
	realTop, err := filepath.EvalSymlinks(top)
	if err != nil {
		f.t.Fatalf("eval top symlinks: %v", err)
	}
	realGitDir, err := filepath.EvalSymlinks(gitDir)
	if err != nil {
		f.t.Fatalf("eval gitdir symlinks: %v", err)
	}
	if realTop != realRepo {
		f.t.Fatalf("git top-level %q != fixture repo %q", realTop, realRepo)
	}
	rel, err := filepath.Rel(realRepo, realGitDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		f.t.Fatalf("git dir %q does not lie inside fixture repo %q", realGitDir, realRepo)
	}
}

func (f *frontDoorFixture) git(args ...string) string {
	f.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = f.repo
	cmd.Env = f.env(nil)
	out, err := cmd.CombinedOutput()
	if err != nil {
		f.t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func (f *frontDoorFixture) writeStubs() {
	f.t.Helper()
	markerScript := `#!/usr/bin/env python3
import json, os, sys
name = sys.argv[1]
args = sys.argv[2:]
stdin_data = ""
if name.startswith("hook-pre-push"):
    stdin_data = sys.stdin.read()
event = {
    "name": name,
    "args": args,
    "stdin": stdin_data,
    "git_dir": os.environ.get("GIT_DIR", ""),
    "git_work_tree": os.environ.get("GIT_WORK_TREE", ""),
    "literal_pathspecs": os.environ.get("GIT_LITERAL_PATHSPECS", ""),
}
with open(os.environ["MARKER_FILE"], "a", encoding="utf-8") as fp:
    fp.write(json.dumps(event) + "\n")
if os.environ.get("FAIL_MARKER") == name:
    sys.exit(7)
`
	sshScript := `#!/bin/sh
interval="${SSH_ALIVE_INTERVAL:-60}"
printf 'serveraliveinterval %s\n' "$interval"
`
	if err := os.WriteFile(filepath.Join(f.bin, "marker"), []byte(markerScript), 0o755); err != nil {
		f.t.Fatalf("write marker stub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(f.bin, "ssh"), []byte(sshScript), 0o755); err != nil {
		f.t.Fatalf("write ssh stub: %v", err)
	}
}

func (f *frontDoorFixture) copyRepoFiles(ciSuiteRefs string) {
	f.t.Helper()
	for _, rel := range []string{"tools/tt", ".githooks/pre-commit", ".githooks/pre-push"} {
		src := filepath.Join(f.root, rel)
		dst := filepath.Join(f.repo, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			f.t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			f.t.Fatalf("read %s: %v", src, err)
		}
		if err := os.WriteFile(dst, data, 0o755); err != nil {
			f.t.Fatalf("write %s: %v", dst, err)
		}
	}
	f.writeJustfile(ciSuiteRefs)
}

func (f *frontDoorFixture) writeJustfile(ciSuiteRefs string) {
	f.t.Helper()
	raw, err := os.ReadFile(filepath.Join(f.root, "justfile"))
	if err != nil {
		f.t.Fatalf("read justfile: %v", err)
	}
	text := string(raw)
	replacements := map[string]string{
		"fast_cmd":          `"marker test-fast"`,
		"test_cmd":          `"marker test"`,
		"one_cmd":           `"marker test-one"`,
		"check_hygiene_cmd": `"marker check-hygiene"`,
		"ci_hygiene_cmd":    `"marker ci-check-hygiene"`,
		"check_code_cmd":    `"marker check-code"`,
		"docs_check_cmd":    `"marker hook-pre-commit-docs"`,
		"ci_suite_refs":     `"` + ciSuiteRefs + `"`,
	}
	for name, val := range replacements {
		re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `\s*:=\s*.*$`)
		text = re.ReplaceAllString(text, name+" := "+val)
	}
	if err := os.WriteFile(filepath.Join(f.repo, "justfile"), []byte(text), 0o644); err != nil {
		f.t.Fatalf("write fixture justfile: %v", err)
	}
}

func (f *frontDoorFixture) setPolicy(content *string) {
	f.t.Helper()
	path := filepath.Join(f.repo, ".tt.toml")
	if content == nil {
		_ = os.Remove(path)
		return
	}
	if err := os.WriteFile(path, []byte(*content), 0o644); err != nil {
		f.t.Fatalf("write .tt.toml: %v", err)
	}
}

func (f *frontDoorFixture) resetLogs() {
	f.t.Helper()
	_ = os.Remove(f.markerFile)
	_ = os.Remove(f.ttLog)
}

func (f *frontDoorFixture) run(extraEnv map[string]string, stdin string, name string, args ...string) (int, string, string) {
	f.t.Helper()
	f.resetLogs()
	cmd := exec.Command(name, args...)
	cmd.Dir = f.repo
	cmd.Env = f.env(extraEnv)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			f.t.Fatalf("exec %s %v: %v", name, args, err)
		}
	}
	return code, stdout.String(), stderr.String()
}

func (f *frontDoorFixture) markers() []markerEvent {
	f.t.Helper()
	data, err := os.ReadFile(f.markerFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		f.t.Fatalf("read markers: %v", err)
	}
	var out []markerEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev markerEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			f.t.Fatalf("unmarshal marker %q: %v", line, err)
		}
		out = append(out, ev)
	}
	return out
}

func (f *frontDoorFixture) markerNames() []string {
	events := f.markers()
	names := make([]string, 0, len(events))
	for _, ev := range events {
		names = append(names, ev.Name)
	}
	return names
}

func (f *frontDoorFixture) ttRecords() []ttRecord {
	f.t.Helper()
	data, err := os.ReadFile(f.ttLog)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		f.t.Fatalf("read ttLog: %v", err)
	}
	var out []ttRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var rec ttRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			f.t.Fatalf("unmarshal ttRecord %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

func assertSliceEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d]: got %q, want %q (full got=%v, want=%v)", label, i, got[i], want[i], got, want)
		}
	}
}

func assertTTRecord(t *testing.T, rec ttRecord, target, execState, reason string, wantExit *int) {
	t.Helper()
	if rec.Target != target || rec.Execution != execState || rec.ExecutionReason != reason {
		t.Fatalf("tt record mismatch: got (%q, %q, %q), want (%q, %q, %q)",
			rec.Target, rec.Execution, rec.ExecutionReason, target, execState, reason)
	}
	if wantExit == nil {
		if rec.Exit != nil {
			t.Fatalf("tt record %s: expected nil exit, got %d", target, *rec.Exit)
		}
		if rec.Tests != nil || rec.Selection != nil || rec.Slowest != nil {
			t.Fatalf("tt record %s: expected nil tests/selection/slowest on skip, got %+v", target, rec)
		}
	} else {
		if rec.Exit == nil || *rec.Exit != *wantExit {
			t.Fatalf("tt record %s: expected exit %d, got %v", target, *wantExit, rec.Exit)
		}
	}
}

func intPtr(v int) *int { return &v }

func TestFrontDoorPolicyOnAndOff(t *testing.T) {
	f := newFrontDoorFixture(t)
	zero := intPtr(0)
	seven := intPtr(7)

	// 1. .tt.toml absent -> execution_reason == "default"
	f.setPolicy(nil)
	for _, tc := range []struct {
		recipe  string
		markers []string
		targets []string
	}{
		{"check-hygiene", []string{"check-hygiene"}, []string{"check-hygiene"}},
		{"ci-check-hygiene", []string{"ci-check-hygiene"}, []string{"ci-check-hygiene"}},
		{"check-code", []string{"check-code"}, []string{"check-code"}},
		{"test-fast", []string{"test-fast"}, []string{"test-fast"}},
		{"test", []string{"test"}, []string{"test"}},
		{"check", []string{"check-hygiene", "check-code"}, []string{"check-hygiene", "check-code"}},
		{"gate", []string{"check-hygiene", "check-code", "test"}, []string{"check-hygiene", "check-code", "test"}},
		{"hook-pre-commit", []string{"check-hygiene", "check-code"}, []string{"hook-pre-commit", "check-code"}},
		{"hook-pre-push", []string{"check-hygiene", "check-code", "test"}, []string{"hook-pre-push", "check-code", "test"}},
		{"hook-pre-push-fast", []string{"check-hygiene", "check-code", "test-fast"}, []string{"hook-pre-push-fast", "check-code", "test-fast"}},
	} {
		code, _, stderr := f.run(nil, "", "just", tc.recipe)
		if code != 0 {
			t.Fatalf("absent policy: just %s exited %d: %s", tc.recipe, code, stderr)
		}
		assertSliceEqual(t, "markers("+tc.recipe+")", f.markerNames(), tc.markers)
		recs := f.ttRecords()
		if len(recs) != len(tc.targets) {
			t.Fatalf("absent policy: just %s recorded %d rows, want %d", tc.recipe, len(recs), len(tc.targets))
		}
		for i, target := range tc.targets {
			assertTTRecord(t, recs[i], target, "ran", "default", zero)
		}
	}

	// test-full under absent policy -> single forced "test" row
	code, _, stderr := f.run(nil, "", "just", "test-full")
	if code != 0 {
		t.Fatalf("absent policy: just test-full exited %d: %s", code, stderr)
	}
	assertSliceEqual(t, "markers(test-full)", f.markerNames(), []string{"test"})
	recs := f.ttRecords()
	if len(recs) != 1 {
		t.Fatalf("absent policy: test-full recorded %d rows, want 1", len(recs))
	}
	assertTTRecord(t, recs[0], "test", "ran", "forced", zero)

	// Short-circuit when check-hygiene fails
	code, _, _ = f.run(map[string]string{"FAIL_MARKER": "check-hygiene"}, "", "just", "gate")
	if code != 7 {
		t.Fatalf("expected gate to exit 7 when check-hygiene fails, got %d", code)
	}
	assertSliceEqual(t, "short-circuit(check-hygiene)", f.markerNames(), []string{"check-hygiene"})

	// Short-circuit when check-code fails
	code, _, _ = f.run(map[string]string{"FAIL_MARKER": "check-code"}, "", "just", "gate")
	if code != 7 {
		t.Fatalf("expected gate to exit 7 when check-code fails, got %d", code)
	}
	assertSliceEqual(t, "short-circuit(check-code)", f.markerNames(), []string{"check-hygiene", "check-code"})

	// 2. .tt.toml with automatic_broad = true -> "policy-enabled" for broad targets, "default" for hygiene/focused
	policyTrue := "automatic_broad = true\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyTrue)
	code, _, stderr = f.run(nil, "", "just", "gate")
	if code != 0 {
		t.Fatalf("automatic_broad=true: just gate exited %d: %s", code, stderr)
	}
	recs = f.ttRecords()
	if len(recs) != 3 {
		t.Fatalf("automatic_broad=true: gate recorded %d rows, want 3", len(recs))
	}
	assertTTRecord(t, recs[0], "check-hygiene", "ran", "default", zero)
	assertTTRecord(t, recs[1], "check-code", "ran", "policy-enabled", zero)
	assertTTRecord(t, recs[2], "test", "ran", "policy-enabled", zero)

	code, _, stderr = f.run(nil, "", "just", "test-one", "./internal/ui")
	if code != 0 {
		t.Fatalf("automatic_broad=true: just test-one exited %d: %s", code, stderr)
	}
	recs = f.ttRecords()
	if len(recs) != 1 {
		t.Fatalf("automatic_broad=true: test-one recorded %d rows, want 1", len(recs))
	}
	assertTTRecord(t, recs[0], "test-one", "ran", "default", zero)

	// 3. .tt.toml with automatic_broad = false -> "policy-disabled" with nil exit for broad targets
	policyFalse := "automatic_broad = false\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyFalse)

	// Focused test-one preserves complex arguments and records "default"
	wantArgs := []string{"./internal/ui", "two words", "it's", "$HOME", "glob *", "-run", "TestFoo"}
	code, _, stderr = f.run(nil, "", "just", append([]string{"test-one"}, wantArgs...)...)
	if code != 0 {
		t.Fatalf("automatic_broad=false: just test-one exited %d: %s", code, stderr)
	}
	events := f.markers()
	if len(events) != 1 || events[0].Name != "test-one" {
		t.Fatalf("expected 1 test-one marker event, got %+v", events)
	}
	assertSliceEqual(t, "test-one args", events[0].Args, wantArgs)
	assertTTRecord(t, f.ttRecords()[0], "test-one", "ran", "default", zero)

	// test-one with no arguments fails
	code, _, _ = f.run(nil, "", "just", "test-one")
	if code == 0 {
		t.Fatalf("expected just test-one with no args to exit nonzero")
	}

	// Composite and direct broad targets skip broad work and record exit: null
	for _, tc := range []struct {
		recipe        string
		wantMarkers   []string
		mandatoryName string
		skippedNames  []string
	}{
		{"check", []string{"check-hygiene"}, "check-hygiene", []string{"check-code"}},
		{"gate", []string{"check-hygiene"}, "check-hygiene", []string{"check-code", "test"}},
		{"hook-pre-commit", []string{"check-hygiene"}, "hook-pre-commit", []string{"check-code"}},
		{"hook-pre-push", []string{"check-hygiene"}, "hook-pre-push", []string{"check-code", "test"}},
		{"hook-pre-push-fast", []string{"check-hygiene"}, "hook-pre-push-fast", []string{"check-code", "test-fast"}},
	} {
		code, _, stderr := f.run(nil, "", "just", tc.recipe)
		if code != 0 {
			t.Fatalf("automatic_broad=false: just %s exited %d: %s", tc.recipe, code, stderr)
		}
		if !strings.Contains(stderr, "skipped by policy") {
			t.Fatalf("automatic_broad=false: just %s stderr missing 'skipped by policy': %q", tc.recipe, stderr)
		}
		assertSliceEqual(t, "policy-off markers("+tc.recipe+")", f.markerNames(), tc.wantMarkers)
		recs := f.ttRecords()
		if len(recs) != 1+len(tc.skippedNames) {
			t.Fatalf("automatic_broad=false: just %s recorded %d rows, want %d", tc.recipe, len(recs), 1+len(tc.skippedNames))
		}
		assertTTRecord(t, recs[0], tc.mandatoryName, "ran", "default", zero)
		for i, sk := range tc.skippedNames {
			assertTTRecord(t, recs[1+i], sk, "skipped", "policy-disabled", nil)
		}
	}

	// Forced-failure fixture: FAIL_MARKER=test under automatic_broad = false
	// Ordinary `just test` skips the failing stub, exits 0, and records exit == nil
	code, _, stderr = f.run(map[string]string{"FAIL_MARKER": "test"}, "", "just", "test")
	if code != 0 {
		t.Fatalf("forced-failure fixture: expected policy-off just test to exit 0 on skip, got %d: %s", code, stderr)
	}
	if len(f.markers()) != 0 {
		t.Fatalf("forced-failure fixture: expected no marker spawned on policy skip, got %v", f.markerNames())
	}
	recs = f.ttRecords()
	if len(recs) != 1 {
		t.Fatalf("forced-failure fixture: expected 1 tt record on skip, got %d", len(recs))
	}
	assertTTRecord(t, recs[0], "test", "skipped", "policy-disabled", nil)

	// Forced runs (`TT_FORCE=1 just test` and `just test-full`) execute the failing stub, exit 7, and record exit == 7
	for _, cmdArgs := range [][]string{{"test"}, {"test-full"}} {
		extra := map[string]string{"FAIL_MARKER": "test"}
		if cmdArgs[0] == "test" {
			extra["TT_FORCE"] = "1"
		}
		code, _, _ := f.run(extra, "", "just", cmdArgs...)
		if code != 7 {
			t.Fatalf("forced-failure fixture: expected %v to exit 7, got %d", cmdArgs, code)
		}
		assertSliceEqual(t, "forced-failure markers", f.markerNames(), []string{"test"})
		recs := f.ttRecords()
		if len(recs) != 1 {
			t.Fatalf("forced-failure fixture: expected 1 tt record for %v, got %d", cmdArgs, len(recs))
		}
		assertTTRecord(t, recs[0], "test", "ran", "forced", seven)
	}
}

func TestFrontDoorHostedBroadTargetsRequireForceAndPropagateFailure(t *testing.T) {
	f := newFrontDoorFixture(t)
	policyFalse := "automatic_broad = false\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyFalse)

	// Without TT_FORCE=1 on GITHUB_ACTIONS=true, direct and composite broad recipes exit 2
	for _, recipe := range []string{"test", "test-fast", "check-code", "check", "gate", "hook-pre-commit", "hook-pre-push", "hook-pre-push-fast"} {
		code, _, stderr := f.run(map[string]string{"GITHUB_ACTIONS": "true"}, "", "just", recipe)
		if code != 2 {
			t.Fatalf("hosted without force: expected just %s to exit 2, got %d (stderr=%s)", recipe, code, stderr)
		}
		if !strings.Contains(stderr, "TT_FORCE=1") {
			t.Fatalf("hosted without force: expected stderr for %s to mention TT_FORCE=1, got %q", recipe, stderr)
		}
		for _, m := range f.markerNames() {
			if m == "test" || m == "test-fast" || m == "check-code" {
				t.Fatalf("hosted without force: just %s spawned broad marker %s", recipe, m)
			}
		}
		for _, r := range f.ttRecords() {
			if r.Execution == "skipped" {
				t.Fatalf("hosted without force: just %s recorded skipped row %+v", recipe, r)
			}
		}
	}

	// With TT_FORCE=1 on GITHUB_ACTIONS=true, gate runs all children and propagates failure
	code, _, _ := f.run(map[string]string{
		"GITHUB_ACTIONS": "true",
		"TT_FORCE":       "1",
		"FAIL_MARKER":    "test",
	}, "", "just", "gate")
	if code != 7 {
		t.Fatalf("hosted with TT_FORCE=1: expected gate to propagate exit 7, got %d", code)
	}
	assertSliceEqual(t, "hosted forced gate markers", f.markerNames(), []string{"check-hygiene", "check-code", "test"})
	recs := f.ttRecords()
	if len(recs) != 3 {
		t.Fatalf("hosted with TT_FORCE=1: expected 3 tt records, got %d", len(recs))
	}
	assertTTRecord(t, recs[0], "check-hygiene", "ran", "forced", intPtr(0))
	assertTTRecord(t, recs[1], "check-code", "ran", "forced", intPtr(0))
	assertTTRecord(t, recs[2], "test", "ran", "forced", intPtr(7))
}

func TestPreCommitHookClassificationAndLiteralPathspecs(t *testing.T) {
	f := newFrontDoorFixture(t)
	hookPath := filepath.Join(f.repo, ".githooks", "pre-commit")

	stageFile := func(rel, content string) {
		full := filepath.Join(f.repo, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
		f.git("add", "--", rel)
	}

	// 1. Staged AGENTS.md routes to hook-pre-commit-docs even with GIT_LITERAL_PATHSPECS=1
	stageFile("AGENTS.md", "# guide\n")
	code, _, stderr := f.run(map[string]string{"GIT_LITERAL_PATHSPECS": "1"}, "", "sh", hookPath)
	if code != 0 {
		t.Fatalf("pre-commit AGENTS.md failed (%d): %s", code, stderr)
	}
	events := f.markers()
	if len(events) != 1 || events[0].Name != "hook-pre-commit-docs" {
		t.Fatalf("expected hook-pre-commit-docs for AGENTS.md, got %+v", events)
	}
	if events[0].Literal != "" {
		t.Fatalf("expected GIT_LITERAL_PATHSPECS to be cleared, got %q", events[0].Literal)
	}

	// 2. Staged README.md routes to hook-pre-commit
	stageFile("README.md", "# readme\n")
	code, _, stderr = f.run(nil, "", "sh", hookPath)
	if code != 0 {
		t.Fatalf("pre-commit README.md failed (%d): %s", code, stderr)
	}
	assertSliceEqual(t, "pre-commit README.md markers", f.markerNames(), []string{"check-hygiene", "check-code"})

	// 3. Rename from code file to AGENTS.md routes to hook-pre-commit
	f.git("reset", "-q")
	stageFile("src.go", "package main\n")
	f.git("commit", "-qm", "initial")
	f.git("mv", "-f", "src.go", "AGENTS.md")
	code, _, stderr = f.run(nil, "", "sh", hookPath)
	if code != 0 {
		t.Fatalf("pre-commit rename failed (%d): %s", code, stderr)
	}
	assertSliceEqual(t, "pre-commit rename markers", f.markerNames(), []string{"check-hygiene", "check-code"})
}

func TestPrePushHookRoutingAndPolicyWarning(t *testing.T) {
	f := newFrontDoorFixture(t)
	hookPath := filepath.Join(f.repo, ".githooks", "pre-push")
	policyFalse := "automatic_broad = false\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyFalse)

	shaA := strings.Repeat("a", 40)
	shaB := strings.Repeat("b", 40)
	refLine := "refs/heads/main " + shaA + " refs/heads/main " + shaB + "\n"
	tagLine := "refs/tags/v1.0.0 " + shaA + " refs/tags/v1.0.0 " + shaB + "\n"
	wantWarn := "pre-push: local test skipped by policy; remote full-suite coverage is not declared for this push. Use TT_FORCE=1 for deliberate local verification."

	// 1. Uncovered push (ci_suite_refs = "") under automatic_broad = false warns, clears GIT vars, and runs hook-pre-push
	f.writeJustfile("")
	code, _, stderr := f.run(map[string]string{
		"GIT_DIR":               filepath.Join(f.repo, ".git"),
		"GIT_WORK_TREE":         f.repo,
		"GIT_LITERAL_PATHSPECS": "1",
	}, refLine, "sh", hookPath, "origin", "https://example.invalid/repo.git")
	if code != 0 {
		t.Fatalf("uncovered pre-push failed (%d): %s", code, stderr)
	}
	if !strings.Contains(stderr, wantWarn) {
		t.Fatalf("expected uncovered pre-push stderr to contain warning %q, got %q", wantWarn, stderr)
	}
	recs := f.ttRecords()
	if len(recs) != 3 || recs[0].Target != "hook-pre-push" {
		t.Fatalf("expected hook-pre-push tt records, got %+v", recs)
	}
	events := f.markers()
	if len(events) != 1 || events[0].GitDir != "" || events[0].GitWorkTree != "" || events[0].Literal != "" {
		t.Fatalf("expected GIT_DIR, GIT_WORK_TREE, GIT_LITERAL_PATHSPECS cleared in hook child, got %+v", events)
	}

	// 2. Covered push (ci_suite_refs = "refs/heads/*" on origin) selects hook-pre-push-fast without warning
	f.writeJustfile("refs/heads/*")
	code, _, stderr = f.run(nil, refLine, "sh", hookPath, "origin", "https://example.invalid/repo.git")
	if code != 0 {
		t.Fatalf("covered pre-push failed (%d): %s", code, stderr)
	}
	if strings.Contains(stderr, wantWarn) {
		t.Fatalf("unexpected uncovered warning on covered push: %q", stderr)
	}
	recs = f.ttRecords()
	if len(recs) != 3 || recs[0].Target != "hook-pre-push-fast" {
		t.Fatalf("expected hook-pre-push-fast tt records, got %+v", recs)
	}

	// 3. Tag push or non-origin remote with ci_suite_refs = "refs/heads/*" falls back to hook-pre-push and warns
	code, _, stderr = f.run(nil, tagLine, "sh", hookPath, "origin", "https://example.invalid/repo.git")
	if code != 0 || !strings.Contains(stderr, wantWarn) {
		t.Fatalf("tag pre-push expected exit 0 with warning, got code=%d stderr=%q", code, stderr)
	}
	if f.ttRecords()[0].Target != "hook-pre-push" {
		t.Fatalf("expected hook-pre-push for tag ref, got %+v", f.ttRecords())
	}

	// 4. Malformed .tt.toml exits 2 before running just
	badPolicy := "[policy]\nautomatic_broad = false\n"
	f.setPolicy(&badPolicy)
	f.writeJustfile("")
	code, _, stderr = f.run(nil, refLine, "sh", hookPath, "origin", "https://example.invalid/repo.git")
	if code != 2 {
		t.Fatalf("expected exit 2 on malformed .tt.toml, got %d (stderr=%q)", code, stderr)
	}
	if !strings.Contains(stderr, "pre-push: cannot inspect test execution policy:") {
		t.Fatalf("expected malformed policy error message, got %q", stderr)
	}
	if len(f.markers()) != 0 {
		t.Fatalf("expected no recipe execution on malformed policy, got %v", f.markerNames())
	}

	// 5. Missing SSH keepalive (serveraliveinterval 0) exits 1 before running just
	f.setPolicy(&policyFalse)
	code, _, stderr = f.run(map[string]string{"SSH_ALIVE_INTERVAL": "0"}, refLine, "sh", hookPath, "origin", "git@example.invalid:khughitt/tasks-tui.git")
	if code != 1 {
		t.Fatalf("expected exit 1 when SSH serveraliveinterval is 0, got %d (stderr=%q)", code, stderr)
	}
	if len(f.markers()) != 0 {
		t.Fatalf("expected no recipe execution when SSH keepalive check fails, got %v", f.markerNames())
	}
}

func TestFrontDoorCleanHomeCIHygienePassesWithoutRegistry(t *testing.T) {
	root := checkoutRoot(t)
	tmpHome, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + tmpHome,
		"XDG_CONFIG_HOME=" + filepath.Join(tmpHome, ".config"),
		"XDG_DATA_HOME=" + filepath.Join(tmpHome, ".local", "share"),
		"XDG_CACHE_HOME=" + filepath.Join(tmpHome, ".cache"),
		"XDG_STATE_HOME=" + filepath.Join(tmpHome, ".local", "state"),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"USER=testuser",
		"TERM=dumb",
		"LANG=C.UTF-8",
	}
	for _, args := range [][]string{
		{"tools/ops-check"},
		{"tools/ops-docs", "check"},
	} {
		cmd := exec.Command("python3", args...)
		cmd.Dir = root
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("python3 %v failed: %v\n%s", args, err, out)
		}
		if len(bytes.TrimSpace(out)) != 0 {
			t.Fatalf("python3 %v unexpected output: %s", args, out)
		}
	}
}

func extractWorkflowStepRun(t *testing.T, workflowText, stepName string) string {
	t.Helper()
	re := regexp.MustCompile(`(?ms)^\s*-\s*name:\s*` + regexp.QuoteMeta(stepName) + `\s*\n\s*run:\s*\|\n((?:\s{10,}.*\n|\s*\n)+)`)
	m := re.FindStringSubmatch(workflowText)
	if m == nil {
		t.Fatalf("step %q run block not found in ci.yml", stepName)
	}
	lines := strings.Split(m[1], "\n")
	var unindented []string
	for _, line := range lines {
		if strings.HasPrefix(line, "          ") {
			unindented = append(unindented, strings.TrimPrefix(line, "          "))
		} else if strings.TrimSpace(line) == "" {
			unindented = append(unindented, "")
		} else {
			break
		}
	}
	return strings.Join(unindented, "\n")
}

func TestFrontDoorCIWorkflowStructureAndPins(t *testing.T) {
	root := checkoutRoot(t)
	wfPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	raw, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("read .github/workflows/ci.yml: %v", err)
	}
	text := string(raw)

	if strings.Contains(text, "pull_request") {
		t.Fatalf("ci.yml must not declare a duplicate pull_request trigger")
	}
	for _, needle := range []string{
		"branches: ['**']",
		"workflow_dispatch:",
		"contents: read",
		"group: ${{ github.workflow }}-${{ github.ref }}",
		"cancel-in-progress: true",
		"runs-on: ubuntu-24.04",
		"timeout-minutes: 15",
		`TT_FORCE: "1"`,
		`GIT_CONFIG_NOSYSTEM: "1"`,
		`TASKS_PINNED_SHA: "c47a2537386dd9cc357a6576445bf5838e0c8c3e"`,
		"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1",
		"fetch-depth: 0",
		"persist-credentials: false",
		"actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e",
		"go-version-file: go.mod",
		"cache: true",
		"extractions/setup-just@53165ef7e734c5c07cb06b3c8e7b647c5aa16db3",
		"just-version: 1.59.0",
		"actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a",
		"retention-days: 7",
		"if-no-files-found: error",
		"$RUNNER_TEMP/tasks-src",
		"$RUNNER_TEMP/tasks-src/target/release",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("ci.yml missing required substring %q", needle)
		}
	}

	stepOrder := []string{
		"- name: Prepare isolated Git config and evidence directory",
		"- name: Checkout",
		"- name: Setup Go",
		"- name: Install just",
		"- name: Build pinned tasks CLI",
		"- name: Record tool versions",
		"- name: Run setup",
		"- name: Run portable hygiene checks",
		"- name: Run Go format, vet, and staticcheck",
		"- name: Run full test suite and verify coverage",
		"- name: Upload CI evidence",
	}
	lastIdx := -1
	for _, step := range stepOrder {
		idx := strings.Index(text, step)
		if idx < 0 {
			t.Fatalf("ci.yml missing step %q", step)
		}
		if idx <= lastIdx {
			t.Fatalf("ci.yml step %q out of order", step)
		}
		lastIdx = idx
	}

	verifyScript := extractWorkflowStepRun(t, text, "Run full test suite and verify coverage")
	cpIdx := strings.Index(verifyScript, `cp "$HOME/.local/share/ops/runs.jsonl" "$EVIDENCE_DIR/runs.jsonl"`)
	exitIdx := strings.Index(verifyScript, `if [ "$rc" -ne 0 ]; then`)
	if cpIdx < 0 || exitIdx < 0 || cpIdx > exitIdx {
		t.Fatalf("ci.yml must copy runs.jsonl into EVIDENCE_DIR before exiting on nonzero rc")
	}
}

func TestFrontDoorCICoverageVerifierEnforcesDynamicPackagesAndUncachedTests(t *testing.T) {
	root := checkoutRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read .github/workflows/ci.yml: %v", err)
	}
	script := extractWorkflowStepRun(t, string(raw), "Run full test suite and verify coverage")

	allPkgs := []string{
		"github.com/khughitt/tasks-tui/cmd/tasks-tui",
		"github.com/khughitt/tasks-tui/internal/config",
		"github.com/khughitt/tasks-tui/internal/identity",
		"github.com/khughitt/tasks-tui/internal/integration",
		"github.com/khughitt/tasks-tui/internal/launch",
		"github.com/khughitt/tasks-tui/internal/tasksctl",
		"github.com/khughitt/tasks-tui/internal/ui",
		"github.com/khughitt/tasks-tui/internal/uitest",
	}

	validLog := strings.Join([]string{
		"=== RUN   TestMainFlag",
		"--- PASS: TestMainFlag (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/cmd/tasks-tui\t0.020s",
		"=== RUN   TestConfigLoad",
		"--- PASS: TestConfigLoad (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/config\t0.015s",
		"=== RUN   TestIdentitySlot",
		"--- PASS: TestIdentitySlot (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/identity\t0.012s",
		"=== RUN   TestProjectViewAgainstRealBinary",
		"--- PASS: TestProjectViewAgainstRealBinary (0.50s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/integration\t0.510s",
		"=== RUN   TestLaunchCommand",
		"--- PASS: TestLaunchCommand (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/launch\t0.010s",
		"=== RUN   TestRunnerList",
		"--- PASS: TestRunnerList (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/tasksctl\t0.025s",
		"=== RUN   TestAppUpdate",
		"=== RUN   TestAppUpdate/subcase",
		"--- PASS: TestAppUpdate (0.02s)",
		"    --- PASS: TestAppUpdate/subcase (0.01s)",
		"ok  \tgithub.com/khughitt/tasks-tui/internal/ui\t0.030s",
		"?   \tgithub.com/khughitt/tasks-tui/internal/uitest\t[no test files]",
	}, "\n") + "\n"

	validRunRow := `{"target":"test","execution":"ran","execution_reason":"forced","exit":0}` + "\n"

	runStep := func(pkgs []string, testLog, runsJSONL string, justExit int) (int, string, string, string) {
		base, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatalf("eval symlinks: %v", err)
		}
		repo := filepath.Join(base, "repo")
		home := filepath.Join(base, "home")
		bin := filepath.Join(base, "bin")
		evidence := filepath.Join(base, "evidence")
		for _, d := range []string{repo, home, bin, evidence} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
		}

		// Initialize isolated git repo with one commit so `git rev-parse HEAD` succeeds
		env := []string{
			"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"),
			"HOME=" + home,
			"EVIDENCE_DIR=" + evidence,
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.invalid",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.invalid",
			"USER=testuser",
			"TERM=dumb",
			"LANG=C.UTF-8",
		}
		for _, gArgs := range [][]string{{"init", "-q"}, {"commit", "--allow-empty", "-qm", "init"}} {
			c := exec.Command("git", gArgs...)
			c.Dir = repo
			c.Env = env
			if out, err := c.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", gArgs, err, out)
			}
		}

		// Stub `go` and `just`
		goStub := "#!/bin/sh\nif [ \"$1\" = \"list\" ] && [ \"$2\" = \"./...\" ]; then\n  cat <<'EOF'\n" + strings.Join(pkgs, "\n") + "\nEOF\n  exit 0\nfi\nexit 1\n"
		if err := os.WriteFile(filepath.Join(bin, "go"), []byte(goStub), 0o755); err != nil {
			t.Fatalf("write go stub: %v", err)
		}
		logFile := filepath.Join(base, "stub-test.log")
		if err := os.WriteFile(logFile, []byte(testLog), 0o644); err != nil {
			t.Fatalf("write stub test log: %v", err)
		}
		runsSource := filepath.Join(base, "stub-runs.jsonl")
		if err := os.WriteFile(runsSource, []byte(runsJSONL), 0o644); err != nil {
			t.Fatalf("write stub runs: %v", err)
		}
		justStub := `#!/bin/sh
if [ "$1" = "test" ]; then
  mkdir -p "$HOME/.local/share/ops"
  if [ -s "` + runsSource + `" ]; then
    cp "` + runsSource + `" "$HOME/.local/share/ops/runs.jsonl"
  fi
  cat "` + logFile + `"
  exit ` + strings.TrimSpace(string(rune('0'+justExit))) + `
fi
exit 1
`
		if err := os.WriteFile(filepath.Join(bin, "just"), []byte(justStub), 0o755); err != nil {
			t.Fatalf("write just stub: %v", err)
		}

		cmd := exec.Command("bash", "-c", script)
		cmd.Dir = repo
		cmd.Env = env
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		code := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = exitErr.ExitCode()
			} else {
				t.Fatalf("run verifier: %v", err)
			}
		}
		return code, stdout.String(), stderr.String(), evidence
	}

	// 1. Positive case
	code, stdout, stderr, evidence := runStep(allPkgs, validLog, validRunRow, 0)
	if code != 0 {
		t.Fatalf("positive verifier case failed (%d):\nstdout=%s\nstderr=%s", code, stdout, stderr)
	}
	summaryBytes, err := os.ReadFile(filepath.Join(evidence, "coverage-summary.json"))
	if err != nil {
		t.Fatalf("read coverage-summary.json: %v", err)
	}
	var summary struct {
		PackageCount    int `json:"package_count"`
		ExecutedTests   int `json:"executed_tests"`
		FailedTests     int `json:"failed_tests"`
		SkippedTests    int `json:"skipped_tests"`
		SkippedModules  int `json:"skipped_modules"`
		ExcludedModules int `json:"excluded_modules"`
	}
	if err := json.Unmarshal(summaryBytes, &summary); err != nil {
		t.Fatalf("unmarshal coverage-summary.json: %v", err)
	}
	if summary.PackageCount != 8 || summary.ExecutedTests != 7 || summary.FailedTests != 0 ||
		summary.SkippedTests != 0 || summary.SkippedModules != 0 || summary.ExcludedModules != 0 {
		t.Fatalf("unexpected coverage summary: %+v", summary)
	}

	// 2. Suite failure preserves runs.jsonl in EVIDENCE_DIR before exiting nonzero
	failRow := `{"target":"test","execution":"ran","execution_reason":"forced","exit":1}` + "\n"
	code, _, _, evidence = runStep(allPkgs, validLog, failRow, 1)
	if code != 1 {
		t.Fatalf("expected exit 1 on suite failure, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(evidence, "runs.jsonl")); err != nil {
		t.Fatalf("expected runs.jsonl copied to EVIDENCE_DIR on failure: %v", err)
	}

	// 3. Negative cases
	negatives := []struct {
		name string
		pkgs []string
		log  string
		runs string
	}{
		{
			name: "missing integration package from output",
			pkgs: allPkgs,
			log:  strings.ReplaceAll(validLog, "ok  \tgithub.com/khughitt/tasks-tui/internal/integration\t0.510s\n", ""),
			runs: validRunRow,
		},
		{
			name: "cached package result",
			pkgs: allPkgs,
			log:  strings.ReplaceAll(validLog, "ok  \tgithub.com/khughitt/tasks-tui/internal/ui\t0.030s", "ok  \tgithub.com/khughitt/tasks-tui/internal/ui\t(cached)"),
			runs: validRunRow,
		},
		{
			name: "skipped test line",
			pkgs: allPkgs,
			log:  strings.ReplaceAll(validLog, "--- PASS: TestProjectViewAgainstRealBinary (0.50s)", "--- SKIP: TestProjectViewAgainstRealBinary (0.00s)"),
			runs: validRunRow,
		},
		{
			name: "failed test line",
			pkgs: allPkgs,
			log:  strings.ReplaceAll(validLog, "--- PASS: TestConfigLoad (0.01s)", "--- FAIL: TestConfigLoad (0.01s)"),
			runs: validRunRow,
		},
		{
			name: "unexpected no test files package",
			pkgs: allPkgs,
			log:  strings.ReplaceAll(validLog, "ok  \tgithub.com/khughitt/tasks-tui/internal/launch\t0.010s", "?   \tgithub.com/khughitt/tasks-tui/internal/launch\t[no test files]"),
			runs: validRunRow,
		},
		{
			name: "zero executed tests",
			pkgs: []string{"github.com/khughitt/tasks-tui/internal/uitest"},
			log:  "?   \tgithub.com/khughitt/tasks-tui/internal/uitest\t[no test files]\n",
			runs: validRunRow,
		},
		{
			name: "missing tt test row",
			pkgs: allPkgs,
			log:  validLog,
			runs: `{"target":"check-code","execution":"ran","execution_reason":"forced","exit":0}` + "\n",
		},
		{
			name: "unforced tt test row",
			pkgs: allPkgs,
			log:  validLog,
			runs: `{"target":"test","execution":"ran","execution_reason":"default","exit":0}` + "\n",
		},
	}
	for _, tc := range negatives {
		code, _, _, _ := runStep(tc.pkgs, tc.log, tc.runs, 0)
		if code == 0 {
			t.Fatalf("negative verifier case %q unexpectedly succeeded", tc.name)
		}
	}
}

func checkRepoPolicyContract(root string, env []string) error {
	evalJust := func(variable string) (string, error) {
		cmd := exec.Command("just", "--evaluate", variable)
		cmd.Dir = root
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
	ciRemote, err := evalJust("ci_remote")
	if err != nil || ciRemote != "origin" {
		return os.ErrInvalid
	}
	ciSuiteRefs, err := evalJust("ci_suite_refs")
	if err != nil {
		return err
	}
	policyBytes, err := os.ReadFile(filepath.Join(root, ".tt.toml"))
	if os.IsNotExist(err) {
		if ciSuiteRefs != "" {
			return os.ErrInvalid
		}
		return nil
	}
	if err != nil {
		return err
	}
	policyText := string(policyBytes)
	if strings.Contains(policyText, "automatic_broad = true") {
		if ciSuiteRefs != "" {
			return os.ErrInvalid
		}
		return nil
	}
	if !strings.Contains(policyText, "automatic_broad = false") ||
		!strings.Contains(policyText, `broad_targets = ["test", "test-fast", "check-code"]`) {
		return os.ErrInvalid
	}
	if ciSuiteRefs != "refs/heads/*" {
		return os.ErrInvalid
	}
	wfBytes, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil || !strings.Contains(string(wfBytes), `TT_FORCE: "1"`) {
		return os.ErrInvalid
	}
	return nil
}

func TestFrontDoorLiveRepoPolicyAndCICoverageContract(t *testing.T) {
	f := newFrontDoorFixture(t)
	if err := checkRepoPolicyContract(f.root, f.env(nil)); err != nil {
		t.Fatalf("live repo failed policy/CI coverage contract: %v", err)
	}

	// Copy ci.yml into fixture repo for activated-state checks
	wfSrc, err := os.ReadFile(filepath.Join(f.root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	wfDst := filepath.Join(f.repo, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(wfDst), 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(wfDst, wfSrc, 0o644); err != nil {
		t.Fatalf("write fixture ci.yml: %v", err)
	}

	// Valid enabled (absent .tt.toml, ci_suite_refs == "")
	f.setPolicy(nil)
	f.writeJustfile("")
	if err := checkRepoPolicyContract(f.repo, f.env(nil)); err != nil {
		t.Fatalf("expected absent .tt.toml + empty ci_suite_refs to pass: %v", err)
	}

	// Valid rolled-back (automatic_broad = true, ci_suite_refs == "")
	policyTrue := "automatic_broad = true\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyTrue)
	f.writeJustfile("")
	if err := checkRepoPolicyContract(f.repo, f.env(nil)); err != nil {
		t.Fatalf("expected automatic_broad=true + empty ci_suite_refs to pass: %v", err)
	}

	// Valid activated (automatic_broad = false, ci_suite_refs == "refs/heads/*")
	policyFalse := "automatic_broad = false\nbroad_targets = [\"test\", \"test-fast\", \"check-code\"]\n"
	f.setPolicy(&policyFalse)
	f.writeJustfile("refs/heads/*")
	if err := checkRepoPolicyContract(f.repo, f.env(nil)); err != nil {
		t.Fatalf("expected automatic_broad=false + refs/heads/* to pass: %v", err)
	}

	// Invalid: automatic_broad = false with empty ci_suite_refs
	f.writeJustfile("")
	if err := checkRepoPolicyContract(f.repo, f.env(nil)); err == nil {
		t.Fatalf("expected automatic_broad=false + empty ci_suite_refs to fail contract")
	}

	// Invalid: automatic_broad = true with ci_suite_refs == "refs/heads/*"
	f.setPolicy(&policyTrue)
	f.writeJustfile("refs/heads/*")
	if err := checkRepoPolicyContract(f.repo, f.env(nil)); err == nil {
		t.Fatalf("expected automatic_broad=true + refs/heads/* to fail contract")
	}
}
