package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tasks-tui/internal/identity"
	"tasks-tui/internal/uitest"
)

type stubView struct {
	name          string
	loaded        int
	loadingNow    bool
	capturingKeys bool
	seen          []string
}

func (s *stubView) title() string   { return s.name }
func (s *stubView) reload() tea.Cmd { s.loaded++; return nil }
func (s *stubView) update(msg tea.Msg) (view, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		s.seen = append(s.seen, k.String())
	}
	return s, nil
}
func (s *stubView) render(_, _ int) string { return "view " + s.name }
func (s *stubView) current() *target       { return nil }
func (s *stubView) project() string        { return "tui" }
func (s *stubView) capturing() bool        { return s.capturingKeys }
func (s *stubView) loading() bool          { return s.loadingNow }

func testApp(stack ...view) *App {
	return New(&Env{Styles: NewStyles(identity.Tone{Dark: true})}, Options{Stack: stack})
}
func TestAppShellKeysStackAndStatus(t *testing.T) {
	root, child := &stubView{name: "root"}, &stubView{name: "child"}
	app := testApp(root)
	d := drive(t, app)
	d.Expect("view root", "root", "? keys")
	d.Feed(pushMsg{v: child})
	d.Expect("view child", "root › child")
	if child.loaded != 1 {
		t.Fatalf("child reloads on push: %d", child.loaded)
	}
	d.Key("esc")
	if root.loaded != 2 {
		t.Fatalf("root reloads on pop: %d", root.loaded)
	}
	d.Key("?")
	d.Expect("quick add", "launch agent")
	d.Key("q")
	if d.Quit {
		t.Fatal("q closes the legend without quitting")
	}
	d.ExpectNot("launch agent")
	d.Feed(tea.WindowSizeMsg{Width: 24, Height: 60})
	d.Key("?")
	for _, line := range strings.Split(d.Screen(), "\n") {
		if lipgloss.Width(line) > 24 {
			t.Fatalf("a narrow tall terminal still clips the legend to its width: %q", line)
		}
	}
	d.Key("esc")
	d.Feed(tea.WindowSizeMsg{Width: 40, Height: 20})
	d.Key("?")
	d.Expect("↓ more")
	d.ExpectNot("quit / close")
	for range 40 {
		d.Key("j")
	}
	d.Expect("quit / close")
	d.ExpectNot("↓ more")
	d.Key("esc")
	d.Feed(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.Notice(LevelError, "boom")
	d.Feed(tickMsg{})
	d.Expect("✗ boom")
	d.Key("j")
	d.ExpectNot("boom")
	app.Notice(LevelWarning, "careful")
	d.Feed(tickMsg{})
	d.Expect("⚠ careful")
	d.Key("j")
	d.Key("W")
	d.Expect("messages", "boom")
	d.Key("esc")
	d.Key("q")
	if !d.Quit {
		t.Fatal("q must quit")
	}
}

func TestCapturingViewOwnsEveryKey(t *testing.T) {
	v := &stubView{name: "root", capturingKeys: true}
	app := testApp(v)
	d := drive(t, app)
	for _, k := range []string{"q", "s", "backspace", "?", "/"} {
		d.Key(k)
	}
	if d.Quit || len(app.stack) != 1 || app.legend || len(v.seen) != 5 {
		t.Fatalf("captured state: quit=%v stack=%d legend=%v seen=%v", d.Quit, len(app.stack), app.legend, v.seen)
	}
}

func TestNoPromptLaunchCopiesRenderedPrompt(t *testing.T) {
	app := testApp(&stubView{name: "root"})
	_, cmd := app.Update(launchMsg{harness: "crush", id: "tui-1", dir: "/wt", prompt: "Run `tasks start tui-1`", noPrompt: true})
	if cmd == nil || fmt.Sprint(cmd()) != "Run `tasks start tui-1`" {
		t.Fatalf("clipboard command = %v", cmd)
	}
	if app.log.Current() == nil || !strings.Contains(app.log.Current().Text, "prompt copied") {
		t.Fatalf("notice = %v", app.log.Current())
	}
}

type loaderView struct {
	name    string
	loader  *Loader
	stored  string
	notices int
}

func (v *loaderView) title() string { return v.name }
func (v *loaderView) reload() tea.Cmd {
	return v.loader.Request(func(gen uint64) tea.Cmd { return v.loader.Cmd(gen, func() (any, error) { return "fresh", nil }) })
}
func (v *loaderView) update(raw tea.Msg) (view, tea.Cmd) {
	msg, ok := raw.(loadMsg)
	if !ok || msg.loader != v.loader.ID() {
		return v, nil
	}
	accept, next := v.loader.Done(msg.gen)
	if !accept || msg.background {
		return v, next
	}
	if msg.err != nil {
		v.notices++
		return v, notices(LevelError, msg.err.Error())
	}
	v.stored = msg.data.(string)
	return v, next
}
func (v *loaderView) render(_, _ int) string { return v.name + ":" + v.stored }
func (v *loaderView) current() *target       { return nil }
func (v *loaderView) project() string        { return "" }
func (v *loaderView) capturing() bool        { return false }
func (v *loaderView) loading() bool          { return v.loader.InFlight() }

func TestStatusLineRightAlignsHintAndSpinsWhileLoading(t *testing.T) {
	v := &stubView{name: "root"}
	app := testApp(v)
	d := drive(t, app)
	status := strings.Split(strings.TrimRight(d.Screen(), "\n"), "\n")
	last := status[len(status)-1]
	if !strings.HasSuffix(last, "? keys") || lipgloss.Width(last) != 120 {
		t.Fatalf("hint right-aligned to the width: %q", last)
	}
	v.loadingNow = true
	if _, cmd := app.Update(tea.KeyPressMsg{Code: 'r', Text: "r"}); cmd == nil || !app.spinning {
		t.Fatal("a loading view starts the spinner")
	}
	app.Update(app.spinner.Tick())
	if !strings.Contains(d.Screen(), app.spinner.View()) {
		t.Fatalf("spinner glyph on the status line:\n%s", d.Screen())
	}
	v.loadingNow = false
	if _, cmd := app.Update(app.spinner.Tick()); cmd != nil || app.spinning {
		t.Fatal("spinner stops when nothing loads")
	}
	if strings.Contains(d.Screen(), app.spinner.View()) {
		t.Fatalf("no spinner when idle:\n%s", d.Screen())
	}
}

func TestStatusLineKeepsHintWhenLeftTextIsLong(t *testing.T) {
	app := testApp(&stubView{name: strings.Repeat("crumb/", 20)})
	d := drive(t, app)
	d.Feed(tea.WindowSizeMsg{Width: 32, Height: 10})
	last := strings.Split(strings.TrimRight(d.Screen(), "\n"), "\n")[9]
	if !strings.HasSuffix(last, "? keys") || lipgloss.Width(last) != 32 {
		t.Fatalf("long breadcrumb hides hint: %q", last)
	}

	app.Notice(LevelError, strings.Repeat("notice ", 20))
	last = strings.Split(strings.TrimRight(d.Screen(), "\n"), "\n")[9]
	if !strings.HasSuffix(last, "? keys") || lipgloss.Width(last) != 32 {
		t.Fatalf("long notice hides hint: %q", last)
	}
}

func TestStatusLineClipsHintAtTinyWidth(t *testing.T) {
	app := testApp(&stubView{name: "root"})
	d := drive(t, app)
	d.Feed(tea.WindowSizeMsg{Width: 3, Height: 5})
	last := strings.Split(strings.TrimRight(d.Screen(), "\n"), "\n")[4]
	if lipgloss.Width(last) != 3 {
		t.Fatalf("tiny status line width: %q", last)
	}
	app.width = 0
	if got := app.statusLine(); got != "" {
		t.Fatalf("zero status line width: %q", got)
	}
}

func TestHiddenLoadSuppressesDataAndNoticesThenReloads(t *testing.T) {
	hidden := &loaderView{name: "hidden", loader: NewLoader()}
	app := testApp(hidden)
	if cmd := hidden.reload(); cmd == nil {
		t.Fatal("initial load must start")
	}
	app.push(&stubView{name: "top"})
	gen := hidden.loader.gen
	app.Update(loadMsg{loader: hidden.loader.ID(), gen: gen, data: "hidden data", err: errors.New("hidden failure")})
	if hidden.stored != "" || hidden.notices != 0 || len(app.Messages()) != 0 {
		t.Fatalf("hidden result leaked: stored=%q notices=%d log=%v", hidden.stored, hidden.notices, app.Messages())
	}
	cmd := app.pop()
	if cmd == nil {
		t.Fatal("hidden result must drain loader bookkeeping so reload can run")
	}
	_, follow := app.Update(cmd())
	if follow != nil {
		follow()
	}
	if hidden.stored != "fresh" {
		t.Fatalf("foreground reload stored %q", hidden.stored)
	}
}

func TestMessagesReturnsLog(t *testing.T) {
	app := testApp(&stubView{name: "root"})
	app.Notice(LevelWarning, "careful")
	if got := app.Messages(); len(got) != 1 || !strings.Contains(got[0].Text, "careful") {
		t.Fatalf("messages: %v", got)
	}
}

// chord presses the prefix through Update so its expiry tick is not run by the driver, then the second key.
func chord(t *testing.T, d *uitest.Driver, app *App, first, second string) {
	t.Helper()
	app.Update(tea.KeyPressMsg{Code: []rune(first)[0], Text: first})
	d.Key(second)
}

func TestChordPrefixWaitsThenDeliversOrExpires(t *testing.T) {
	v := &stubView{name: "root"}
	app := testApp(v)
	d := drive(t, app)
	if _, cmd := app.Update(tea.KeyPressMsg{Code: 'c', Text: "c"}); app.chord.prefix != "c" || cmd == nil {
		t.Fatalf("c pends a chord and schedules its expiry: %+v", app.chord)
	}
	d.Expect("c …")
	d.Key("c")
	if app.chord.prefix != "" || len(v.seen) != 1 || v.seen[0] != "c c" {
		t.Fatalf("the second key completes the chord as one key press: %v %+v", v.seen, app.chord)
	}
	chord(t, d, app, "c", "x")
	if app.chord.prefix != "" || len(v.seen) != 2 || v.seen[1] != "c x" {
		t.Fatalf("a non-matching second key clears the prefix; the view ignores %q: %v", "c x", v.seen)
	}
	app.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	d.Feed(chordExpireMsg{deadline: time.Time{}})
	if app.chord.prefix != "c" {
		t.Fatal("a stale expiry must not clear a newer prefix")
	}
	d.Feed(chordExpireMsg{deadline: app.chord.deadline})
	if app.chord.prefix != "" {
		t.Fatal("the matching expiry clears the prefix")
	}
	d.Expect("? keys")
}
