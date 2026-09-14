package identity

import "testing"

// Ported verbatim from familiar test/identity.test.js (spec §8.1).
func TestNormalizeRemoteCollapsesForms(t *testing.T) {
	const want = "github.com/me/api"
	for _, in := range []string{
		"git@github.com:me/api.git",
		"https://github.com/me/api.git",
		"https://github.com/me/api",
		"https://user:token@github.com/me/api",
		"https://user:token@github.com/me/api.git",
		"ssh://git@github.com/me/api.git",
		"ssh://git@github.com:22/me/api.git",
		"https://github.com/me/api/",
		"HTTPS://GitHub.com/Me/API.git",
		"git@GitHub.com:Me/API.git",
	} {
		got, ok := NormalizeRemote(in)
		if !ok || got != want {
			t.Errorf("%q -> %q,%v", in, got, ok)
		}
	}
	for in, want := range map[string]string{
		"git@gitlab.example.com:group/subgroup/project.git":     "gitlab.example.com/group/subgroup/project",
		"https://gitlab.example.com/group/subgroup/project.git": "gitlab.example.com/group/subgroup/project",
	} {
		if got, _ := NormalizeRemote(in); got != want {
			t.Errorf("%q -> %q", in, got)
		}
	}
}

func TestNormalizeRemoteScpOwnerVersusSSHPort(t *testing.T) {
	scp, _ := NormalizeRemote("git@github.com:1234/repo.git")
	ssh, _ := NormalizeRemote("ssh://git@github.com:1234/repo.git")
	if scp != "github.com/1234/repo" || ssh != "github.com/repo" {
		t.Fatalf("scp=%q ssh=%q", scp, ssh)
	}
}

func TestNormalizeRemoteRejects(t *testing.T) {
	for _, in := range []string{"", "   ", "/local/path/to/repo"} {
		if _, ok := NormalizeRemote(in); ok {
			t.Errorf("%q should not normalize", in)
		}
	}
}

// Vectors computed with familiar's fnv1a32 (src/protocol/hash.js) on 2026-09-13.
func TestFNV1a32Vectors(t *testing.T) {
	for in, want := range map[string]uint32{
		"":                           2166136261,
		"a":                          3826002220,
		"github.com/khughitt/tasks":  856432771,
		"/home/user/repos/tasks-tui": 880204061,
	} {
		if got := FNV1a32(in); got != want {
			t.Errorf("fnv1a32(%q) = %d, want %d", in, got, want)
		}
	}
	if AutoSlot("github.com/khughitt/tasks") != 7 || AutoSlot("") != 1 {
		t.Fatal("AutoSlot must be fnv1a32 mod 12")
	}
}

func TestSlotTableIsFrozen(t *testing.T) {
	if len(Slots) != SlotCount {
		t.Fatalf("%d hue rows for %d slots", len(Slots), SlotCount)
	}
	if Slots[0] != (HueSat{22, 62}) || Slots[11] != (HueSat{0, 6}) {
		t.Fatalf("table drifted: %v", Slots)
	}
}
