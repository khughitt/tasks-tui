package ui

import "charm.land/bubbles/v2/key"

// chordPrefixes are the first keys of two-key sequences. App.key holds one pending
// prefix and delivers the completed chord to the normal dispatch as a key press whose
// String() is the sequence, e.g. "g g", so bindings list chords as ordinary keys.
var chordPrefixes = []string{"c", "s", "S", "g"}

// launchBinding is one c-chord: the harness it launches, or "" for the picker. The
// second key is fixed here, not derived from harness names (three defaults start with c).
type launchBinding struct{ seq, harness string }

var launchKeys = []launchBinding{{"c l", "claude"}, {"c o", "codex"}, {"c r", "crush"}, {"c O", "opencode"}, {"c c", ""}}

// sortBinding is one s/S-chord in one view: the column and direction it dispatches.
// The conformance test reads column and desc from here, never from the key.
type sortBinding struct {
	seq, column string
	desc        bool
}

var projectsSort = []sortBinding{{"s p", "prefix", false}, {"S p", "prefix", true}, {"s a", "activity", false}, {"S a", "activity", true}}
var tasksSort = []sortBinding{{"s p", "priority", false}, {"S p", "priority", true}, {"s a", "age", false}, {"S a", "age", true}, {"s t", "title", false}, {"S t", "title", true}}

func findSort(table []sortBinding, seq string) (sortBinding, bool) {
	for _, b := range table {
		if b.seq == seq {
			return b, true
		}
	}
	return sortBinding{}, false
}
func launchSeqs() []string {
	out := make([]string, len(launchKeys))
	for i, b := range launchKeys {
		out[i] = b.seq
	}
	return out
}
func sortSeqs() []string {
	seen := map[string]bool{}
	var out []string
	for _, b := range append(append([]sortBinding{}, projectsSort...), tasksSort...) {
		if !seen[b.seq] {
			seen[b.seq] = true
			out = append(out, b.seq)
		}
	}
	return out
}

// scopedKey records where a binding is matched (global in App.key, projects in
// projectsView.update, tasks in projectView.update and taskView.update) and what it does.
// Every keymap field below is built through scoped, so the list and the dispatch table
// are one source; TestEveryKeymapFieldIsScoped fails on a field that bypasses it.
type scopedKey struct {
	scope, action, arg string
	b                  key.Binding
}

var scopedKeys []scopedKey

func scoped(scope, action, arg string, b key.Binding) key.Binding {
	scopedKeys = append(scopedKeys, scopedKey{scope, action, arg, b})
	return b
}

// both registers one binding in the projects and the tasks scope.
func both(action string, b key.Binding) key.Binding {
	return scoped("tasks", action, "", scoped("projects", action, "", b))
}

// localAliases are keys tasks-tui adds to a shared action beyond the vocabulary's own
// aliases; the inventory marks them alias = true.
var localAliases = map[string]bool{"i": true}

type keymap struct {
	Quit, Help, Log, Reload, Back, Add, Launch, Start, Park, Done, Drop, Enter, Filter key.Binding
	Up, Down, Prev, Next, PageUp, PageDown, Top, Bottom                                key.Binding
	Tab, ShiftTab, Digits, Sort, Copy                                                  key.Binding
}

var keys = keymap{
	Quit:     scoped("global", "quit", "", key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit / close"))),
	Help:     scoped("global", "help", "", key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "keys"))),
	Log:      scoped("global", "log", "", key.NewBinding(key.WithKeys("W"), key.WithHelp("W", "messages"))),
	Reload:   scoped("global", "reload", "", key.NewBinding(key.WithKeys("f5"), key.WithHelp("F5", "reload"))),
	Back:     scoped("global", "dismiss", "", key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back"))),
	Add:      scoped("global", "add", "", key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "quick add"))),
	Launch:   key.NewBinding(key.WithKeys(launchSeqs()...), key.WithHelp("c …", "launch agent")), // rows: launchKeys
	Start:    scoped("global", "primary", "start the focused task", key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "start"))),
	Park:     scoped("global", "park", "", key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "park"))),
	Done:     scoped("global", "done", "", key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "done"))),
	Drop:     scoped("global", "remove", "", key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "drop"))),
	Enter:    both("open", key.NewBinding(key.WithKeys("enter", "i"), key.WithHelp("enter/i", "open"))),
	Filter:   both("search", key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))),
	Up:       both("up", key.NewBinding(key.WithKeys("k", "up"))),
	Down:     both("down", key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/k", "move"))),
	Prev:     scoped("tasks", "prev", "", key.NewBinding(key.WithKeys("h", "left"))),
	Next:     scoped("tasks", "next", "", key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("h/l", "prev / next tab"))),
	PageUp:   scoped("tasks", "page", "", key.NewBinding(key.WithKeys("pgup", "ctrl+b"))),
	PageDown: scoped("tasks", "page", "", key.NewBinding(key.WithKeys("pgdown", "ctrl+f"), key.WithHelp("pgup/pgdn", "page"))),
	Top:      scoped("tasks", "top", "", key.NewBinding(key.WithKeys("g g", "home"), key.WithHelp("g g/G", "top / bottom"))),
	Bottom:   scoped("tasks", "bottom", "", key.NewBinding(key.WithKeys("G", "end"))),
	Tab:      scoped("tasks", "tabs", "", key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab/shift+tab", "tabs"))),
	ShiftTab: scoped("tasks", "tabs", "", key.NewBinding(key.WithKeys("shift+tab"))),
	Digits:   scoped("tasks", "select", "", key.NewBinding(key.WithKeys("1", "2", "3", "4", "5"), key.WithHelp("1–5", "tab n"))),
	Sort:     key.NewBinding(key.WithKeys(sortSeqs()...), key.WithHelp("s …/S …", "sort asc / desc")), // rows: projectsSort, tasksSort
	Copy:     scoped("tasks", "yank", "", key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy id"))),
}

// legendGroups is spec v1.1 §9: the panel renders these bindings' Help(); a grouped
// binding without help text fails TestEveryGroupedBindingHasHelp.
var legendGroups = []legendGroup{
	{"move", []key.Binding{keys.Down, keys.Next, keys.Top, keys.PageDown, keys.Tab, keys.Digits}},
	{"open", []key.Binding{keys.Enter, keys.Back, keys.Filter, keys.Sort, keys.Reload}},
	{"task", []key.Binding{keys.Add, keys.Launch, keys.Start, keys.Park, keys.Done, keys.Drop, keys.Copy}},
	{"app", []key.Binding{keys.Help, keys.Log, keys.Quit}},
}
