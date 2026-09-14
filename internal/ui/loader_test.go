package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestLoaderCoalescesAndRejectsStale(t *testing.T) {
	var l Loader
	var issued []uint64
	run := func(gen uint64) tea.Cmd {
		issued = append(issued, gen)
		return func() tea.Msg { return nil }
	}
	if cmd := l.Request(run); cmd == nil {
		t.Fatal("first request runs")
	}
	if cmd := l.Request(run); cmd != nil {
		t.Fatal("second request while in flight is pending, not run")
	}
	if cmd := l.Request(run); cmd != nil {
		t.Fatal("third request coalesces into the one pending")
	}
	if len(issued) != 1 || issued[0] != 1 {
		t.Fatalf("issued %v", issued)
	}
	accept, next := l.Done(1)
	if accept {
		t.Fatal("gen 1 is stale once gen 3 was requested")
	}
	if next == nil || len(issued) != 2 || issued[1] != 3 {
		t.Fatalf("pending reload must run at the latest gen: issued %v", issued)
	}
	accept, next = l.Done(3)
	if !accept || next != nil {
		t.Fatal("the latest result is accepted and nothing is pending")
	}
	if cmd := l.Request(run); cmd == nil {
		t.Fatal("after Done the loader is free again")
	}
}

func TestLoaderIDsAreUnique(t *testing.T) {
	a, b := NewLoader(), NewLoader()
	if a.ID() == b.ID() {
		t.Fatal("loaders must be distinguishable in loadMsg")
	}
}
