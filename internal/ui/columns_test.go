package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var testTable = table{gap: 2, flexMin: 10, cols: []column{
	{key: "gutter", min: 2},
	{key: "id", label: "id", min: 2},
	{key: "n", label: "n", align: alignRight, min: 1, drop: 2},
	{key: "opt", label: "option", min: 6, drop: 1},
	{key: "title", label: "title", min: 5, flex: true},
}}

func plain(s string) cell { return text(lipgloss.NewStyle(), s) }

func testRows() [][]cell {
	return [][]cell{
		{plain("  "), plain("tui-aaa111"), plain("7"), plain("x"), {spans: []span{{"A short title", lipgloss.NewStyle()}}, tail: []span{{"[t]", lipgloss.NewStyle()}}}},
		{plain("  "), plain("t-1"), plain("123"), plain("longer"), {spans: []span{{"A much longer title that will not fit in the column", lipgloss.NewStyle()}}, tail: []span{{"[a, b]", lipgloss.NewStyle()}, {"30d", lipgloss.NewStyle()}}}},
	}
}

func TestWidthsMeasureEveryRowAndKeepLabelMinimum(t *testing.T) {
	w := testTable.widths(testRows(), 200)
	if w[0] != 2 || w[1] != 10 || w[2] != 3 || w[3] != 6 {
		t.Fatalf("widths %v: gutter 2, id widest 10, n widest 3, opt label minimum 6", w)
	}
	if w[4] != 200-(2+10+3+6)-4*2 {
		t.Fatalf("flex takes the remainder: %d", w[4])
	}
}

func TestColumnsDropInOrderThenFlexShrinksToZero(t *testing.T) {
	rows := testRows()
	full := testTable.widths(rows, 39)
	if full[2] == 0 || full[3] == 0 {
		t.Fatalf("39 cells fits everything: %v", full)
	}
	noN := testTable.widths(rows, 34)
	if noN[2] != 0 || noN[3] == 0 {
		t.Fatalf("n (drop 2) goes before opt (drop 1): %v", noN)
	}
	bare := testTable.widths(rows, 26)
	if bare[2] != 0 || bare[3] != 0 || bare[4] != 10 {
		t.Fatalf("both dropped, flex at flexMin: %v", bare)
	}
	tiny := testTable.widths(rows, 12)
	if tiny[4] != 0 {
		t.Fatalf("below minWidth the flexible column is zero, not negative: %v", tiny)
	}
	if got := testTable.minWidth(rows); got != 26 {
		t.Fatalf("minWidth = %d, want 26", got)
	}
}

func TestRowAlignsPadsAndOmitsDroppedColumns(t *testing.T) {
	rows := testRows()
	w := testTable.widths(rows, 60)
	base := lipgloss.NewStyle()
	head := ansi.Strip(testTable.header(w, NewStyles(darkTone)))
	r0 := ansi.Strip(testTable.row(w, rows[0], base))
	r1 := ansi.Strip(testTable.row(w, rows[1], base))
	if !strings.HasPrefix(head, "  "+"  "+"id"+strings.Repeat(" ", 8)+"  "+"  n") {
		t.Fatalf("header keeps blank gutter and right-aligns n: %q", head)
	}
	off := testTable.offset(w, "title")
	if r0[off:off+13] != "A short title" || !strings.HasPrefix(r1[off:], "A much longer") {
		t.Fatalf("title column at offset %d:\n%q\n%q", off, r0, r1)
	}
	nOff := testTable.offset(w, "n")
	if r0[nOff:nOff+3] != "  7" || r1[nOff:nOff+3] != "123" {
		t.Fatalf("right alignment: %q / %q", r0[nOff:nOff+3], r1[nOff:nOff+3])
	}
	dropped := ansi.Strip(testTable.row(testTable.widths(rows, 26), rows[1], base))
	if strings.Contains(dropped, "longer") || strings.Contains(dropped, "123") {
		t.Fatalf("dropped columns render nothing: %q", dropped)
	}
	if lipgloss.Width(dropped) > 26 {
		t.Fatalf("row wider than its table: %d", lipgloss.Width(dropped))
	}
}

func TestFlexTruncatesHeadBeforeTail(t *testing.T) {
	rows := testRows()
	w := testTable.widths(rows, 60)
	r1 := ansi.Strip(testTable.row(w, rows[1], lipgloss.NewStyle()))
	title := r1[testTable.offset(w, "title"):]
	if !strings.HasSuffix(title, "[a, b]  30d") {
		t.Fatalf("tail survives while head can shrink toward flexMin: %q", title)
	}
	if !strings.Contains(title, "…") || lipgloss.Width(title) > 31 {
		t.Fatalf("head truncated with an ellipsis inside the column: %q", title)
	}
	w = testTable.widths(rows, 42)
	title = ansi.Strip(testTable.row(w, rows[1], lipgloss.NewStyle()))[testTable.offset(w, "title"):]
	if lipgloss.Width(title) > 13 || !strings.HasPrefix(title, "A much lo…") {
		t.Fatalf("head keeps flexMin and the tail truncates from its end: %q", title)
	}
}

func TestRowAppliesBaseUnderEverySpanAndFitPads(t *testing.T) {
	base := lipgloss.NewStyle().Background(lipgloss.Color("#202020"))
	cells := []cell{plain("  "), plain("id"), plain("1"), plain("o"), {spans: []span{{"t", lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))}}}}
	w := testTable.widths([][]cell{cells}, 40)
	out := testTable.row(w, cells, base)
	if strings.Count(out, "48;2;32;32;32") < 5 {
		t.Fatalf("background must be under every span and gap: %q", out)
	}
	fitted := fit("abc", 6, base)
	if ansi.Strip(fitted) != "abc   " || !strings.Contains(fitted, "48;2;32;32;32") {
		t.Fatalf("fit pads with base: %q", fitted)
	}
	if got := ansi.Strip(fit("abcdefgh", 4, base)); got != "abcd" {
		t.Fatalf("fit clips: %q", got)
	}
}

func TestColumnMinimumsCoverLabels(t *testing.T) {
	for _, tb := range []table{testTable} {
		for _, c := range tb.cols {
			if c.min < 1 || c.min < lipgloss.Width(c.label) {
				t.Errorf("%s: min %d under label %q", c.key, c.min, c.label)
			}
		}
	}
}
