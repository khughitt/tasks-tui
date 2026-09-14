package ui

import "testing"

func TestLogPrecedenceAndDismiss(t *testing.T) {
	var l Log
	if l.Current() != nil {
		t.Fatal("empty log has no current")
	}
	l.Add(LevelError, "claimed: held")
	l.AddWarnings("park tui-1", []string{"claim store cleanup failed"})
	if c := l.Current(); c == nil || c.Level != LevelError {
		t.Fatalf("a warning must not displace an error: %+v", c)
	}
	if len(l.Entries) != 2 || l.Entries[1].Text != "park tui-1: claim store cleanup failed" {
		t.Fatalf("entries %+v", l.Entries)
	}
	l.Dismiss()
	if l.Current() != nil || len(l.Entries) != 2 {
		t.Fatal("dismiss hides the status line and keeps the log")
	}
	l.Add(LevelWarning, "w")
	l.Add(LevelError, "e")
	if l.Current().Text != "e" {
		t.Fatal("an error displaces a warning")
	}
	l.AddWarnings("x", nil)
	if l.Current().Text != "e" || len(l.Entries) != 4 {
		t.Fatal("no warnings adds nothing")
	}
}
