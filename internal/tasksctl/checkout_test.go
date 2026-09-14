package tasksctl

import "testing"

func TestCheckoutFor(t *testing.T) {
	exists := func(p string) bool { return p == "/wt/park" || p == "/wt/claim" }
	park := &ParkInfo{Worktree: "/wt/park"}
	missingPark := &ParkInfo{Worktree: "/wt/gone-park"}
	live := &ClaimInfo{Worktree: "/wt/claim", Live: true}
	missingLive := &ClaimInfo{Worktree: "/wt/gone-claim", Live: true}
	stale := &ClaimInfo{Worktree: "/wt/gone-stale", Live: false}

	cases := []struct {
		name   string
		park   *ParkInfo
		claim  *ClaimInfo
		dir    string
		notice string
	}{
		{"park wins", park, live, "/wt/park", ""},
		{"missing park falls to live claim", missingPark, live, "/wt/claim", "parked checkout /wt/gone-park is gone; using /wt/claim"},
		{"missing park and live claim fall to root", missingPark, missingLive, "/root", "parked checkout /wt/gone-park is gone; claimed checkout /wt/gone-claim is gone; using /root"},
		{"live claim only", nil, live, "/wt/claim", ""},
		{"missing live claim falls to root", nil, missingLive, "/root", "claimed checkout /wt/gone-claim is gone; using /root"},
		{"stale claim is ignored", nil, stale, "/root", ""},
		{"neither", nil, nil, "/root", ""},
	}
	for _, c := range cases {
		got := CheckoutFor(c.park, c.claim, "/root", exists)
		if got.Dir != c.dir || got.Notice != c.notice {
			t.Fatalf("%s: got %+v", c.name, got)
		}
	}
}
