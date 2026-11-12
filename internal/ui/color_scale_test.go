package ui

import (
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

func TestCountPosition(t *testing.T) {
	for _, tc := range []struct {
		n, maximum int
		want       float64
	}{
		{0, 0, 0}, {1, 1, 0.430676558073393}, {2, 4, 0.682606194485985}, {4, 4, 1}, {239, 239, 1},
	} {
		if got := countPosition(tc.n, tc.maximum); math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("%d/%d: %.15f, want %.15f", tc.n, tc.maximum, got, tc.want)
		}
	}
	previous := 0.0
	for _, n := range []int{1, 2, 4, 39, 239, 1000000} {
		got := countPosition(n, 1000000)
		if got <= previous || got > 1 {
			t.Fatalf("not monotone/bounded: %d %.6f", n, got)
		}
		previous = got
	}
}

func TestRecencyPosition(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		age  time.Duration
		want float64
	}{{0, 1}, {-time.Hour, 1}, {30 * 24 * time.Hour, 0}, {31 * 24 * time.Hour, 0}} {
		got, ok := recencyPosition(now.Add(-tc.age).Format(time.RFC3339), now)
		if !ok || math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("age %s: %g, %v", tc.age, got, ok)
		}
	}
	previous := 1.0
	for _, age := range []time.Duration{time.Minute, time.Hour, 24*time.Hour - time.Minute, 24 * time.Hour, 24*time.Hour + time.Minute, 7 * 24 * time.Hour} {
		got, ok := recencyPosition(now.Add(-age).Format(time.RFC3339), now)
		if !ok || got >= previous || got < 0 {
			t.Fatalf("age %s not monotone: %g", age, got)
		}
		previous = got
	}
	if _, ok := recencyPosition("invalid", now); ok {
		t.Fatal("invalid timestamp accepted")
	}
}

func TestRampContrast(t *testing.T) {
	// Independent reference values pin the sRGB gamma and luminance coefficients.
	if got := contrastRatio(lipgloss.Color("#777777"), color.Black); math.Abs(got-4.68949989000882) > 1e-9 {
		t.Fatalf("gray reference: %.12f", got)
	}
	if got := contrastRatio(lipgloss.Color("#ff0000"), color.Black); math.Abs(got-5.252) > 1e-9 {
		t.Fatalf("red reference: %.12f", got)
	}
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	for _, theme := range []struct{ bg, fg string }{{"#101010", "#eeeeee"}, {"#ffffff", "#171717"}, {"#ffffff", "#777777"}, {"#777777", "#888888"}} {
		s := coloredStyles()
		s.background = lipgloss.Color(theme.bg)
		s.foreground = lipgloss.Color(theme.fg)
		for _, selected := range []bool{false, true} {
			bg := s.rowBackground("tui", selected)
			for _, n := range []int{1, 2, 4, 39, 239} {
				c := s.countStyle(n, 239, bg).GetForeground()
				if got := contrastRatio(c, bg); got < 4.5 {
					t.Fatalf("%s count %d selected=%v: %.8f", theme.bg, n, selected, got)
				}
			}
			floor := math.Min(4.5, contrastRatio(s.foreground, bg))
			for _, age := range []time.Duration{0, time.Minute, time.Hour, 24 * time.Hour, 30 * 24 * time.Hour} {
				c := s.ageStyle(now.Add(-age).Format(time.RFC3339), now, bg).GetForeground()
				if got := contrastRatio(c, bg); got+1e-10 < floor {
					t.Fatalf("date floor %.8f < %.8f", got, floor)
				}
			}
			old := s.ageStyle(now.Add(-30*24*time.Hour).Format(time.RFC3339), now, bg).GetForeground()
			or, og, ob, _ := old.RGBA()
			fr, fg, fb, _ := s.foreground.RGBA()
			if or != fr || og != fg || ob != fb {
				t.Fatal("old date endpoint did not preserve terminal foreground")
			}
		}
	}
}

func quantitativeView(t *testing.T) (*projectsView, time.Time) {
	t.Helper()
	e := testEnv(newFake())
	e.Styles = coloredStyles()
	v := newProjectsView(e)
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	stamp := now.Format(time.RFC3339)
	rows := []tasksctl.Project{
		{Prefix: "tui", Root: "/r/visible-name", Reachable: true, Counts: &tasksctl.Counts{Doing: 1, Todo: 239, Idea: 1, Blocked: 1}, LastActivity: &stamp},
		{Prefix: "ops", Root: "/r/ops", Reachable: true, Counts: &tasksctl.Counts{Doing: 4, Todo: 1000000, Idea: 39}, LastActivity: &stamp},
		{Prefix: "gone", Root: "/r/gone", Reachable: false, Counts: &tasksctl.Counts{Doing: 1000000}},
	}
	v.main.Request(func(uint64) tea.Cmd { return nil })
	v.update(loadMsg{loader: v.main.ID(), gen: v.main.gen, data: projectsData{projects: tasksctl.ProjectsResult{Projects: rows}}})
	return v, now
}

func TestHomeQuantitativeColors(t *testing.T) {
	v, now := quantitativeView(t)
	p := v.data.projects.Projects[0]
	before := v.projectCells(p, now, false)
	if before[1].spans[0].style.Render("1") == before[3].spans[0].style.Render("1") {
		t.Fatal("independent column domains collapsed")
	}
	if !sgr(before[4].spans[0].style.Render("1"), "31") {
		t.Fatal("blocked lost ANSI red")
	}
	v.filter = "tui"
	v.sortAndFilter()
	after := v.projectCells(p, now, false)
	if before[1].spans[0].style.Render("1") != after[1].spans[0].style.Render("1") {
		t.Fatal("filter renormalized counts")
	}
	if v.maxima.Doing != 4 {
		t.Fatalf("unreachable count changed domain: %+v", v.maxima)
	}
	s := v.env.Styles
	if got := s.ageStyle(now.Format(time.RFC3339), now, s.background).Render("now"); !strings.Contains(got, "38;2;57;184;193") {
		t.Fatalf("recent date should be cyan: %q", got)
	}
	if got := s.countStyle(4, 4, s.background).Render("4"); !strings.Contains(got, "38;2;233;180;76") {
		t.Fatalf("large count should be gold: %q", got)
	}
	s.foreground = nil
	if out := s.ageStyle(now.Format(time.RFC3339), now, s.background).Render("now"); out != "now" {
		t.Fatalf("missing foreground must defer date shading: %q", out)
	}
	if out := s.countStyle(1, 4, s.background).Render("1"); !strings.Contains(out, "38;2;") {
		t.Fatal("background alone should enable counts")
	}
	s.setProfile(colorprofile.ANSI)
	if out := s.countStyle(1, 4, s.background).Render("1"); out != "1" {
		t.Fatal("basic profile should have ordinary numeric foreground")
	}
}

func TestCompactHomeLayout(t *testing.T) {
	v, now := quantitativeView(t)
	for _, c := range projectsTable.cols {
		if c.key == "name" {
			t.Fatal("Name still reserves table width")
		}
	}
	screen := ansi.Strip(v.renderAt(100, 12, now))
	if strings.Contains(screen, "visible-name") || strings.Contains(screen, " name ") {
		t.Fatalf("Name still visible: %s", screen)
	}
	v.filter = "visible-name"
	v.sortAndFilter()
	if len(v.rows) != 1 || v.rows[0].Prefix != "tui" {
		t.Fatal("hidden names stopped matching the filter")
	}
	for _, line := range strings.Split(v.renderAt(30, 12, now), "\n") {
		if lipgloss.Width(line) > 30 {
			t.Fatalf("narrow overflow: %q", line)
		}
	}
}
