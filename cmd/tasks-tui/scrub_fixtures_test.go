package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestFixtureScrubber(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		n    int
		meta bool
	}{
		{"color", 1, false}, {"overflow", 39, false}, {"metadata", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			var projects []map[string]any
			for i := range tc.n {
				prefix := fmt.Sprintf("real%02d", i)
				p := map[string]any{"prefix": prefix, "root": "/private/checkouts/" + prefix, "color": "#123AbC", "color_256": 42}
				if tc.meta {
					p["records"] = "/private/records/" + prefix
					p["checkout"] = "/private/checkouts/" + prefix
					p["name"] = "Private project title"
					p["purpose"] = "Private project purpose"
					p["retired"] = map[string]any{"date": "2026-11-01", "successor": prefix, "reason": "old project"}
				}
				projects = append(projects, p)
			}
			raw, err := json.Marshal(map[string]any{"projects": projects, "warnings": []string{}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "registry.json"), raw, 0o600); err != nil {
				t.Fatal(err)
			}
			stub := "#!" + python + "\nfrom pathlib import Path\nprint(Path(__file__).with_name('registry.json').read_text())\n"
			if err := os.WriteFile(filepath.Join(dir, "tasks"), []byte(stub), 0o700); err != nil {
				t.Fatal(err)
			}
			run := func(input []byte) ([]byte, error) {
				cmd := exec.Command(python, filepath.Join(checkoutRoot(t), "tools", "scrub-fixtures"))
				cmd.Env = []string{"PATH=" + dir + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + dir,
					"XDG_CONFIG_HOME=" + filepath.Join(dir, "config"), "XDG_STATE_HOME=" + filepath.Join(dir, "state"), "USER=fixture", "LANG=C.UTF-8"}
				cmd.Stdin = bytes.NewReader(input)
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if err != nil {
					return nil, fmt.Errorf("%w: %s", err, stderr.String())
				}
				return stdout.Bytes(), nil
			}
			out, err := run(raw)
			if err != nil {
				t.Fatal(err)
			}
			again, err := run(raw)
			if err != nil || !bytes.Equal(out, again) {
				t.Fatalf("capture must be deterministic: %v", err)
			}
			if bytes.Contains(out, []byte("/private/")) || bytes.Contains(out, []byte("Private project")) {
				t.Fatalf("private metadata survived: %s", out)
			}
			var doc struct{ Projects []map[string]any }
			if err := json.Unmarshal(out, &doc); err != nil {
				t.Fatal(err)
			}
			seen := map[string]bool{}
			valid := regexp.MustCompile(`^[a-z][a-z0-9]{1,7}$`)
			for i, p := range doc.Projects {
				prefix := p["prefix"].(string)
				if !valid.MatchString(prefix) || seen[prefix] || strings.HasPrefix(prefix, "real") {
					t.Fatalf("unsafe/duplicate synthetic prefix: %q", prefix)
				}
				seen[prefix] = true
				if i == 0 && prefix != "alpha" || i == 25 && prefix != "zulu" {
					t.Fatalf("existing NATO aliases changed: %d %s", i, prefix)
				}
				if p["color"] != "#123AbC" || p["color_256"] != float64(42) {
					t.Fatalf("color values must survive: %+v", p)
				}
			}
			if _, err := run([]byte(`{"unknown":"must refuse"}`)); err == nil || !strings.Contains(err.Error(), "unclassified") {
				t.Fatalf("unknown strings must fail closed: %v", err)
			}
		})
	}
}
