package ui

import "charm.land/bubbles/v2/key"

type keymap struct {
	Quit, Help, Log, Reload, Back, Add, Launch, Start, Park, Done, Drop, Enter, Filter key.Binding
	Up, Down, PageUp, PageDown, Top, Bottom                                            key.Binding
	Tab, ShiftTab, Sort, Copy                                                          key.Binding
}

var keys = keymap{
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "keys")),
	Log:    key.NewBinding(key.WithKeys("W"), key.WithHelp("W", "messages")),
	Reload: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload")),
	Back:   key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "quick add")),
	Launch: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "launch agent")),
	Start:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
	Park:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "park")),
	Done:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "done")),
	Drop:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "drop")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Up:     key.NewBinding(key.WithKeys("k", "up")), Down: key.NewBinding(key.WithKeys("j", "down")),
	PageUp: key.NewBinding(key.WithKeys("pgup", "ctrl+b")), PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+f")),
	Top: key.NewBinding(key.WithKeys("g", "home")), Bottom: key.NewBinding(key.WithKeys("G", "end")),
	Tab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")), ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
	Sort: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sort")), Copy: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy id")),
}

const legendText = `  j/k move   enter open   esc back   tab tabs   / filter   r reload
  a quick add   l launch agent   s start   p park   d done   x drop
  y copy id   W messages   S sort (projects)   ? this legend   q quit`
