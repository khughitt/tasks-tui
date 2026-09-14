// Package uitest drives a Bubble Tea model synchronously for tests.
package uitest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

type Driver struct {
	tb   testing.TB
	m    tea.Model
	Quit bool
}

func New(tb testing.TB, m tea.Model, width, height int) *Driver {
	d := &Driver{tb: tb, m: m}
	d.Feed(tea.WindowSizeMsg{Width: width, Height: height})
	d.Run(m.Init())
	return d
}

func (d *Driver) Run(cmd tea.Cmd) {
	if cmd != nil {
		d.Feed(cmd())
	}
}

func (d *Driver) Feed(msg tea.Msg) {
	switch m := msg.(type) {
	case nil:
		return
	case tea.BatchMsg:
		for _, cmd := range m {
			d.Run(cmd)
		}
		return
	case tea.QuitMsg:
		d.Quit = true
		return
	}
	var cmd tea.Cmd
	d.m, cmd = d.m.Update(msg)
	d.Run(cmd)
}

func (d *Driver) Key(k string) {
	switch k {
	case "enter":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyEnter})
	case "esc":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyEscape})
	case "tab":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyTab})
	case "shift+tab":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	case "backspace":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyBackspace})
	case "ctrl+u":
		d.Feed(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	case "ctrl+r":
		d.Feed(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	case "up":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyUp})
	case "down":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyDown})
	case "pgdown":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyPgDown})
	case "pgup":
		d.Feed(tea.KeyPressMsg{Code: tea.KeyPgUp})
	default:
		r := []rune(k)
		if len(r) != 1 {
			d.tb.Fatalf("Key: %q is not one key; use Type", k)
		}
		d.Feed(tea.KeyPressMsg{Code: r[0], Text: k})
	}
}

func (d *Driver) Type(s string) {
	for _, r := range s {
		d.Feed(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}
func (d *Driver) Screen() string { return ansi.Strip(d.m.View().Content) }
func (d *Driver) Expect(texts ...string) {
	d.tb.Helper()
	screen := d.Screen()
	for _, text := range texts {
		if !strings.Contains(screen, text) {
			d.tb.Fatalf("screen lacks %q:\n%s", text, screen)
		}
	}
}
func (d *Driver) ExpectNot(text string) {
	d.tb.Helper()
	if screen := d.Screen(); strings.Contains(screen, text) {
		d.tb.Fatalf("screen must not contain %q:\n%s", text, screen)
	}
}
