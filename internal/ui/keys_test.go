package ui

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/BurntSushi/toml"
)

// keyRow is one inventory row of tools/keys.toml (ops keys.toml, vendored byte-identical).
type keyRow struct {
	Scope  string `toml:"scope"`
	Seq    string `toml:"seq"`
	Action string `toml:"action"`
	Arg    string `toml:"arg"`
	Alias  bool   `toml:"alias"`
}
type keysFile struct {
	Project map[string]struct {
		Binding []keyRow `toml:"binding"`
	} `toml:"project"`
}

// canonicalNames translates bubbles key spellings to the inventory's; single characters,
// chords ("g g") and ctrl+ combinations are already canonical.
var canonicalNames = map[string]string{
	"esc": "Escape", "enter": "Enter", "backspace": "Backspace", "tab": "Tab", "shift+tab": "shift+Tab",
	"f1": "F1", "f5": "F5", "up": "Up", "down": "Down", "left": "Left", "right": "Right",
	"home": "Home", "end": "End", "pgup": "PgUp", "pgdown": "PgDn", " ": "space",
}

func canonical(k string) string {
	if c, ok := canonicalNames[k]; ok {
		return c
	}
	return k
}

// liveRows is the application's shortcut table: every key of every scoped binding, plus
// the launch and sort chords with the argument each dispatches.
func liveRows() []keyRow {
	var rows []keyRow
	for _, s := range scopedKeys {
		for _, k := range s.b.Keys() {
			r := keyRow{Scope: s.scope, Seq: canonical(k), Action: s.action, Arg: s.arg, Alias: localAliases[k]}
			if s.action == "select" {
				r.Arg = fmt.Sprintf("tab %d", tabFor(k)+1)
			}
			rows = append(rows, r)
		}
	}
	for _, b := range launchKeys {
		arg := b.harness
		if arg == "" {
			arg = "picker"
		}
		rows = append(rows, keyRow{Scope: "global", Seq: b.seq, Action: "launch", Arg: arg})
	}
	for scope, table := range map[string][]sortBinding{"projects": projectsSort, "tasks": tasksSort} {
		for _, b := range table {
			dir := "asc"
			if b.desc {
				dir = "desc"
			}
			rows = append(rows, keyRow{Scope: scope, Seq: b.seq, Action: "sort", Arg: b.column + " " + dir})
		}
	}
	return rows
}

// unscopedFields names the keymap fields whose bindings were not built through scoped
// and are not the two table-driven ones; such a field dispatches keys the inventory
// cannot see.
func unscopedFields(km keymap) []string {
	v := reflect.ValueOf(km)
	var out []string
	for i := 0; i < v.NumField(); i++ {
		name := v.Type().Field(i).Name
		if name == "Launch" || name == "Sort" {
			continue
		}
		b := v.Field(i).Interface().(key.Binding)
		if !slices.ContainsFunc(scopedKeys, func(s scopedKey) bool { return slices.Equal(s.b.Keys(), b.Keys()) }) {
			out = append(out, name)
		}
	}
	return out
}

func TestEveryKeymapFieldIsScoped(t *testing.T) {
	if missing := unscopedFields(keys); len(missing) > 0 {
		t.Fatalf("keymap fields not built through scoped(): %v", missing)
	}
	rogue := keys
	rogue.Add = key.NewBinding(key.WithKeys("+"), key.WithHelp("+", "add"))
	if got := unscopedFields(rogue); !slices.Equal(got, []string{"Add"}) {
		t.Fatalf("a binding that bypasses scoped() must be reported, got %v", got)
	}
}

// inventoryDiff is the conformance comparison: rows in the inventory but not live, and
// rows live but not in the inventory.
func inventoryDiff(inventory, live []keyRow) (missing, extra []string) {
	want, got := map[keyRow]bool{}, map[keyRow]bool{}
	for _, r := range inventory {
		want[r] = true
	}
	for _, r := range live {
		got[r] = true
	}
	for r := range want {
		if !got[r] {
			missing = append(missing, fmt.Sprintf("%+v", r))
		}
	}
	for r := range got {
		if !want[r] {
			extra = append(extra, fmt.Sprintf("%+v", r))
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

func TestLiveBindingsMatchVendoredInventory(t *testing.T) {
	var file keysFile
	if _, err := toml.DecodeFile("../../tools/keys.toml", &file); err != nil {
		t.Fatal(err)
	}
	inventory := file.Project["tasks-tui"].Binding
	if len(inventory) == 0 {
		t.Fatal("tools/keys.toml has no tasks-tui rows")
	}
	live := liveRows()
	for _, r := range live {
		if p, _, ok := strings.Cut(r.Seq, " "); ok && !slices.Contains(chordPrefixes, p) {
			t.Errorf("chord %q has no pending prefix in chordPrefixes", r.Seq)
		}
	}
	if missing, extra := inventoryDiff(inventory, live); len(missing)+len(extra) > 0 {
		t.Fatalf("inventory rows not bound live:\n  %s\nlive bindings not in tools/keys.toml:\n  %s",
			strings.Join(missing, "\n  "), strings.Join(extra, "\n  "))
	}
	// A binding the inventory lacks is reported on the live side, never swallowed.
	unrecorded := keyRow{Scope: "global", Seq: "+", Action: "add", Alias: true}
	if _, extra := inventoryDiff(inventory, append(live, unrecorded)); !slices.Equal(extra, []string{fmt.Sprintf("%+v", unrecorded)}) {
		t.Fatalf("an unrecorded live binding must fail conformance, got %v", extra)
	}
}
