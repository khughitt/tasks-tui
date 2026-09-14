package identity

import (
	"os"
	"path/filepath"
	"testing"
)

const pinsYAML = `
identities:
  - remote: GitHub.com/Me/API
    slot: 3
  - path: ~/d/tasks
    slot: 2
  - path: relative/tasks
    slot: 4
  - path: /nowhere/not-cloned
    slot: 9
  - project: dotfiles
    slot: 11
`

func TestParsePinsAndPrecedence(t *testing.T) {
	pins, err := ParsePins([]byte(pinsYAML))
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	real := filepath.Join(home, "real-tasks")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(home, "d", "tasks")); err != nil {
		t.Fatal(err)
	}
	realpath := filepath.EvalSymlinks

	if slot, ok := MatchPin(pins, "github.com/me/api", real, "tasks", realpath); !ok || slot != 3 {
		t.Fatalf("remote pin: %d %v", slot, ok)
	}
	if slot, ok := MatchPin(pins, "", real, "real-tasks", realpath); !ok || slot != 2 {
		t.Fatalf("path pin via symlink: %d %v", slot, ok)
	}
	work := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	relativeRoot := filepath.Join(work, "relative", "tasks")
	if err := os.MkdirAll(relativeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if slot, ok := MatchPin(pins, "", relativeRoot, "tasks", realpath); !ok || slot != 4 {
		t.Fatalf("relative path pin: %d %v", slot, ok)
	}
	if slot, ok := MatchPin(pins, "", "/elsewhere/dotfiles", "dotfiles", realpath); !ok || slot != 11 {
		t.Fatalf("project pin: %d %v", slot, ok)
	}
	if _, ok := MatchPin(pins, "", "/other", "other", realpath); ok {
		t.Fatal("nothing should match")
	}
}

func TestParsePinsRejectsEmptyPin(t *testing.T) {
	if _, err := ParsePins([]byte("identities:\n  - slot: 1\n")); err == nil {
		t.Fatal("a pin without remote/path/project must error")
	}
	if _, err := ParsePins([]byte("identities:\n  - path: /x\n    slot: 12\n")); err == nil {
		t.Fatal("slot 12 must error")
	}
	if _, err := ParsePins([]byte("identities: 3\n")); err == nil {
		t.Fatal("identities must be a list")
	}
}

func TestLoadPinsMissingIsEmpty(t *testing.T) {
	pins, err := LoadPins(filepath.Join(t.TempDir(), "none.yaml"))
	if err != nil || len(pins.Identities) != 0 {
		t.Fatalf("pins=%+v err=%v", pins, err)
	}
}

func TestSlotForFallsBackToHash(t *testing.T) {
	root := t.TempDir()
	noRemote := func(string) (string, error) { return "", nil }
	slot, err := SlotFor(root, Pins{}, noRemote)
	if err != nil {
		t.Fatal(err)
	}
	canon, _ := filepath.EvalSymlinks(root)
	if slot != AutoSlot(canon) {
		t.Fatalf("slot %d != hash of canonical root %d", slot, AutoSlot(canon))
	}
	withRemote := func(string) (string, error) { return "git@github.com:khughitt/tasks.git", nil }
	slot, err = SlotFor(root, Pins{}, withRemote)
	if err != nil {
		t.Fatal(err)
	}
	if slot != 7 {
		t.Fatalf("remote key slot %d, want 7", slot)
	}
}
