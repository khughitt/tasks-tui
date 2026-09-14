package ui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tasks-tui/internal/identity"
)

type stubView struct {
	name          string
	loaded        int
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

func testApp(stack ...view) *App {
	return New(&Env{Styles: NewStyles(identity.Tone{Dark: true})}, Options{Stack: stack})
}
func TestAppShellKeysStackAndStatus(t *testing.T) {
	root, child := &stubView{name: "root"}, &stubView{name: "child"}
	app := testApp(root)
	d := drive(t, app)
	d.Expect("view root", "root   ? keys")
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
	d.Key("?")
	d.ExpectNot("launch agent")
	app.Notice(LevelError, "boom")
	d.Feed(tickMsg{})
	d.Expect("boom")
	d.Key("j")
	d.ExpectNot("boom")
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
