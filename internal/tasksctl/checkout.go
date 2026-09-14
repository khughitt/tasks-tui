package tasksctl

import "strings"

// Checkout is where a task's commands run and where launch opens (spec §4.1).
type Checkout struct {
	Dir    string
	Notice string // non-empty when a recorded worktree was missing and the rule fell through
}

// CheckoutFor applies §4.1: the park's worktree, else a live claim's worktree, else root.
// exists is injected so the policy is testable without a filesystem.
func CheckoutFor(park *ParkInfo, claim *ClaimInfo, root string, exists func(string) bool) Checkout {
	var missing []string
	if park != nil {
		if exists(park.Worktree) {
			return Checkout{Dir: park.Worktree}
		}
		missing = append(missing, "parked checkout "+park.Worktree+" is gone")
	}
	if claim = liveClaim(claim); claim != nil {
		if exists(claim.Worktree) {
			return Checkout{Dir: claim.Worktree, Notice: checkoutNotice(missing, claim.Worktree)}
		}
		missing = append(missing, "claimed checkout "+claim.Worktree+" is gone")
	}
	return Checkout{Dir: root, Notice: checkoutNotice(missing, root)}
}

func checkoutNotice(missing []string, dir string) string {
	if len(missing) == 0 {
		return ""
	}
	return strings.Join(missing, "; ") + "; using " + dir
}
