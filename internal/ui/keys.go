package ui

import "charm.land/bubbles/v2/key"

// chordPrefixes are the first keys of two-key sequences. App.key holds one pending
// prefix and delivers the completed chord to the normal dispatch as a key press whose
// String() is the sequence, e.g. "c c", so bindings list chords as ordinary keys.
// Only c until the rebinding step frees s, S, and g from their single-key meanings.
var chordPrefixes = []string{"c"}

type keymap struct {
	Quit, Help, Log, Reload, Back, Add, Launch, Start, Park, Done, Drop, Enter, Filter key.Binding
	Up, Down, PageUp, PageDown, Top, Bottom                                            key.Binding
	Tab, ShiftTab, Digits, Sort, Copy                                                  key.Binding
}

var keys = keymap{
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit / close")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "keys")),
	Log:      key.NewBinding(key.WithKeys("W"), key.WithHelp("W", "messages")),
	Reload:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload")),
	Back:     key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "quick add")),
	Launch:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "launch agent")),
	Start:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
	Park:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "park")),
	Done:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "done")),
	Drop:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "drop")),
	Enter:    key.NewBinding(key.WithKeys("enter", "i"), key.WithHelp("enter/i", "open")),
	Filter:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Up:       key.NewBinding(key.WithKeys("k", "up")),
	Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/k", "move")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+b")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+f"), key.WithHelp("pgup/pgdn", "page")),
	Top:      key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g/G", "top / bottom")),
	Bottom:   key.NewBinding(key.WithKeys("G", "end")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab/shift+tab", "tabs")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
	Digits:   key.NewBinding(key.WithKeys("1", "2", "3", "4", "5"), key.WithHelp("1–5", "tab n")),
	Sort:     key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sort")),
	Copy:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy id")),
}

// legendGroups is spec v1.1 §9: the panel renders these bindings' Help(); a grouped
// binding without help text fails TestEveryGroupedBindingHasHelp.
var legendGroups = []legendGroup{
	{"move", []key.Binding{keys.Down, keys.Top, keys.PageDown, keys.Tab, keys.Digits}},
	{"open", []key.Binding{keys.Enter, keys.Back, keys.Filter, keys.Reload}},
	{"task", []key.Binding{keys.Add, keys.Launch, keys.Start, keys.Park, keys.Done, keys.Drop, keys.Copy}},
	{"app", []key.Binding{keys.Help, keys.Log, keys.Sort, keys.Quit}},
}
