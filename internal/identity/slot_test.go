package identity

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitOriginURLDistinguishesMissingOriginFromFailures(t *testing.T) {
	root := t.TempDir()
	if err := exec.Command("git", "init", "--quiet", root).Run(); err != nil {
		t.Fatal(err)
	}
	remote, err := GitOriginURL(root)
	if err != nil || remote != "" {
		t.Fatalf("missing origin: remote=%q err=%v", remote, err)
	}

	path := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	if _, err := GitOriginURL(root); err == nil || !strings.Contains(err.Error(), root) {
		t.Fatal("missing git must error")
	}
	t.Setenv("PATH", path)
	missingRoot := filepath.Join(root, "missing")
	if _, err := GitOriginURL(missingRoot); err == nil || !strings.Contains(err.Error(), missingRoot) {
		t.Fatal("invalid root must error")
	}
}
