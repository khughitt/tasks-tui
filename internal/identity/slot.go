package identity

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitRemote returns origin's URL for a checkout, "" when it has none.
type GitRemote func(root string) (string, error)

// GitOriginURL asks git, with familiar's 2 s bound: a wedged mount must not stall the UI.
func GitOriginURL(root string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", root, "config", "--get", "remote.origin.url").Output()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// SlotFor is spec §8.1: pins by familiar's precedence, else fnv1a32(remote ?? root) mod 12.
func SlotFor(root string, pins Pins, git GitRemote) (int, error) {
	url, err := git(root)
	if err != nil {
		return 0, err
	}
	remote, _ := NormalizeRemote(url)
	repoRoot := canonical(root, filepath.EvalSymlinks)
	if slot, ok := MatchPin(pins, remote, root, filepath.Base(root), filepath.EvalSymlinks); ok {
		return slot, nil
	}
	if remote != "" {
		return AutoSlot(remote), nil
	}
	return AutoSlot(repoRoot), nil
}
