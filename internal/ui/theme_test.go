package ui

import (
	"image/color"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

func seedColors(e *Env) {
	rgb, index := "#123abc", 42
	e.SetProjects([]tasksctl.Project{{Prefix: "tui", Root: "/r/tui", Reachable: true, Color: &rgb, Color256: &index},
		{Prefix: "ops", Root: "/r/ops", Reachable: true, Color: &rgb, Color256: &index}})
}

func coloredStyles() *Styles {
	e := &Env{Styles: NewStyles()}
	seedColors(e)
	e.Styles.setProfile(colorprofile.TrueColor)
	e.Styles.background = color.RGBA{R: 16, G: 16, B: 16, A: 255}
	e.Styles.foreground = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	return e.Styles
}

func sgr(s string, code string) bool {
	for _, match := range regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatch(s, -1) {
		fields := strings.Split(match[1], ";")
		for i := 0; i < len(fields); i++ {
			item := fields[i]
			if (item == "38" || item == "48" || item == "58") && i+1 < len(fields) {
				if fields[i+1] == "2" {
					i += 4
				} else if fields[i+1] == "5" {
					i += 2
				}
				continue
			}
			if item == code {
				return true
			}
		}
	}
	return false
}

func TestProjectStylesFollowProfile(t *testing.T) {
	e := &Env{Styles: NewStyles()}
	seedColors(e)
	for _, tc := range []struct {
		p    colorprofile.Profile
		want string
	}{
		{colorprofile.Unknown, ""}, {colorprofile.TrueColor, "38;2;18;58;188"},
		{colorprofile.ANSI256, "38;5;42"}, {colorprofile.ANSI, ""},
		{colorprofile.ASCII, ""}, {colorprofile.NoTTY, ""},
	} {
		e.Styles.setProfile(tc.p)
		out := e.Styles.Accent("tui").Render("project")
		if tc.want != "" && !strings.Contains(out, tc.want) || tc.want == "" && out != "project" {
			t.Fatalf("%s: %q", tc.p, out)
		}
	}
	rgb := "#123abc"
	e.SetProjects([]tasksctl.Project{{Prefix: "tui", Root: "/r/tui", Color: &rgb}})
	e.Styles.setProfile(colorprofile.ANSI256)
	if out := e.Styles.Accent("tui").Render("x"); out != "x" {
		t.Fatalf("RGB-only must be neutral in indexed mode: %q", out)
	}
}

func TestStartupProjectColors(t *testing.T) {
	e := testEnv(newFake())
	seedColors(e)
	e.Styles.setProfile(colorprofile.TrueColor)
	if out := newProjectView(e, "tui").render(120, 20); !strings.Contains(out, "38;2;18;58;188") {
		t.Fatalf("direct project entry lacks seeded color: %q", out)
	}
	v := newTaskView(e, target{ID: "tui-aaa111", Prefix: "tui"})
	v.data = &taskData{res: tasksctl.ShowResult{Task: tasksctl.Task{ID: "tui-aaa111", Title: "Title", Status: "doing"},
		Parent: &tasksctl.Relation{ID: "ops-bbb222", Title: "Other", Status: "todo"}}}
	if out := v.content(); strings.Count(out, "38;2;18;58;188") < 2 {
		t.Fatalf("direct task/relation colors: %q", out)
	}
	pv := newProjectsView(e)
	gen := pv.main.gen
	pv.main.Request(func(uint64) tea.Cmd { return nil })
	gen = pv.main.gen
	pv.update(loadMsg{loader: pv.main.ID(), gen: gen, data: projectsData{projects: tasksctl.ProjectsResult{
		Projects: []tasksctl.Project{{Prefix: "tui", Root: "/r/tui", Reachable: true}},
	}}})
	if out := e.Styles.Accent("tui").Render("x"); out != "x" {
		t.Fatalf("accepted reload did not remove old color: %q", out)
	}
}

func TestProjectsSelectionWithoutTint(t *testing.T) {
	for _, profile := range []colorprofile.Profile{colorprofile.Unknown, colorprofile.ASCII, colorprofile.ANSI, colorprofile.ANSI256, colorprofile.TrueColor} {
		t.Run(profile.String(), func(t *testing.T) {
			e := testEnv(newFake())
			seedColors(e)
			e.Styles.setProfile(profile)
			v := newProjectsView(e)
			v.rows = []tasksctl.Project{{Prefix: "tui", Root: "/r/tui", Reachable: true, Counts: &tasksctl.Counts{}}, {Prefix: "gone", Root: "/r/gone"}}
			first := ansi.Strip(v.render(100, 10))
			if !strings.Contains(first, "▸ tui") || !strings.Contains(first, "▌ gone") {
				t.Fatalf("initial selection missing: %s", first)
			}
			v.sel = 1
			second := ansi.Strip(v.render(100, 10))
			if !strings.Contains(second, "▌ tui") || !strings.Contains(second, "▸ gone") {
				t.Fatalf("unreachable selection missing: %s", second)
			}
			if lipgloss.Width("▸ ") != 2 || lipgloss.Width("▌ ") != 2 {
				t.Fatal("selection must not change marker width")
			}
		})
	}
}

func TestActiveLabelsWithoutSurface(t *testing.T) {
	s := coloredStyles()
	if out := s.ActiveLabel("tui", "1 Ready"); !sgr(out, "4") || !sgr(out, "1") || strings.Contains(out, "48;") {
		t.Fatalf("colored active tab must use underline, not fill: %q", out)
	}
	s.setProfile(colorprofile.ASCII)
	o := &pickerOverlay{styles: s, prefix: "tui", title: "launch", items: []string{"claude", "codex"}}
	if out := ansi.Strip(o.render(80)); !strings.Contains(out, "[1 claude]") || strings.Contains(out, "[2 codex]") {
		t.Fatalf("monochrome selection missing: %s", out)
	}
	o.update(tea.KeyPressMsg{Code: 'j'})
	if out := ansi.Strip(o.render(80)); !strings.Contains(out, "[2 codex]") || strings.Contains(out, "[1 claude]") {
		t.Fatalf("picker selection did not move: %s", out)
	}
}

func themeApp() *App {
	e := testEnv(newFake())
	return New(e, Options{Stack: []view{newProjectsView(e)}})
}

func commandMessages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	m := cmd()
	if batch, ok := m.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, child := range batch {
			out = append(out, commandMessages(child)...)
		}
		return out
	}
	return []tea.Msg{m}
}

func TestThemeQueryLifecycle(t *testing.T) {
	initial := themeApp()
	initial.opts.ColorQueryTimeout = time.Nanosecond
	for _, msg := range commandMessages(initial.Init()) {
		_, timeout := msg.(colorQueryTimeoutMsg)
		if timeout || reflect.DeepEqual(msg, tea.RequestForegroundColor()) || reflect.DeepEqual(msg, tea.RequestBackgroundColor()) {
			t.Fatalf("Init must not query or start the color timer: %T", msg)
		}
	}
	for _, p := range []colorprofile.Profile{colorprofile.Unknown, colorprofile.ANSI, colorprofile.ASCII, colorprofile.NoTTY} {
		a := themeApp()
		cmd, handled := a.updateTheme(tea.ColorProfileMsg{Profile: p})
		if !handled || len(commandMessages(cmd)) != 0 || len(a.Messages()) != 0 {
			t.Fatalf("basic profile should be quiet: %s", p)
		}
	}
	a := themeApp()
	cmd, _ := a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	var fg, bg int
	messages := commandMessages(cmd)
	for _, msg := range messages {
		if reflect.DeepEqual(msg, tea.RequestForegroundColor()) {
			fg++
		}
		if reflect.DeepEqual(msg, tea.RequestBackgroundColor()) {
			bg++
		}
	}
	if fg != 1 || bg != 1 {
		t.Fatalf("query pair: fg=%d bg=%d", fg, bg)
	}
	if len(messages) != 2 {
		t.Fatalf("zero timeout must produce only the query pair: %d messages", len(messages))
	}
	cmd, _ = a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	if len(commandMessages(cmd)) != 0 {
		t.Fatal("profile repeated the queries")
	}
	a.opts.ColorQueryTimeout = time.Second
	a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.ANSI256})
	generation := a.colorGeneration
	a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.ANSI})
	a.updateTheme(colorQueryTimeoutMsg{generation: generation})
	if len(a.Messages()) != 0 {
		t.Fatal("stale timeout posted in basic mode")
	}
	a.updateTheme(tea.ForegroundColorMsg{Color: color.White})
	a.updateTheme(tea.BackgroundColorMsg{Color: color.Black})
	a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	a.updateTheme(colorQueryTimeoutMsg{generation: a.colorGeneration})
	if len(a.Messages()) != 0 {
		t.Fatal("complete late replies should prevent timeout notice")
	}
	b := themeApp()
	b.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	b.updateTheme(colorQueryTimeoutMsg{generation: b.colorGeneration})
	b.updateTheme(colorQueryTimeoutMsg{generation: b.colorGeneration})
	if len(b.Messages()) != 1 {
		t.Fatalf("one notice per app: %+v", b.Messages())
	}
}

func TestLiveForegroundAndStatusRoles(t *testing.T) {
	s := coloredStyles()
	for name, style := range map[string]lipgloss.Style{"info": s.Info, "tag": s.Tag, "agent": s.ParkAgent, "idea": s.StatusStyle("idea")} {
		if out := style.Render("live"); sgr(out, "2") || strings.Contains(out, "38;") {
			t.Fatalf("%s live text muted: %q", name, out)
		}
	}
	for _, style := range []lipgloss.Style{s.Muted, s.PriorityStyle(4), s.StatusStyle("done"), s.StatusStyle("shelved")} {
		out := spansText([]span{{"inactive", style}}, s.Bold)
		if !sgr(out, "2") || sgr(out, "1") {
			t.Fatalf("dim must not inherit selected bold: %q", out)
		}
	}
	if out := s.Error.Render("blocked"); !sgr(out, "31") || strings.Contains(out, "38;") {
		t.Fatalf("not basic ANSI red: %q", out)
	}
}

func TestMarkdownProfileMarkers(t *testing.T) {
	e := testEnv(newFake())
	v := newTaskView(e, target{})
	for _, p := range []colorprofile.Profile{colorprofile.Unknown, colorprofile.ASCII, colorprofile.NoTTY, colorprofile.ANSI, colorprofile.TrueColor} {
		e.Styles.setProfile(p)
		out, err := v.markdown("**strong** and *emphasis* and ~~gone~~ and `code` and [link](https://example.invalid)")
		if err != nil {
			t.Fatal(err)
		}
		plain := ansi.Strip(out)
		markers := strings.Contains(plain, "**strong**") && strings.Contains(plain, "*emphasis*") && strings.Contains(plain, "~~gone~~") && strings.Contains(plain, "`code`")
		if p <= colorprofile.ASCII && !markers || p >= colorprofile.ANSI && (markers || !sgr(out, "1") || !sgr(out, "3") || !sgr(out, "4")) {
			t.Fatalf("%s wrong emphasis/markers: %q", p, out)
		}
		if strings.Contains(out, "38;") || strings.Contains(out, "48;") {
			t.Fatalf("Markdown forced colors: %q", out)
		}
	}
}

func TestFormAndPromptThemeChanges(t *testing.T) {
	e := testEnv(newFake())
	v := newAddFormView(e, "tui")
	v.titleInput.SetValue("Keep this title")
	v.body.SetValue("Keep this body")
	v.focusField(1)
	v.submitting = true
	p := newPrompt(e.Styles, "prompt", "empty")
	p.input.SetValue("keep prompt")
	a := New(e, Options{Stack: []view{v}})
	a.overlay = p
	a.updateTheme(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	a.updateTheme(tea.BackgroundColorMsg{Color: color.White})
	if v.titleInput.Value() != "Keep this title" || v.body.Value() != "Keep this body" || v.focus != 1 || !v.submitting || p.input.Value() != "keep prompt" {
		t.Fatal("theme change destroyed form state")
	}
}
