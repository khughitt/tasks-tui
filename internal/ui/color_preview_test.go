package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

func TestColorRenderingPreview(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	for _, theme := range []struct {
		name, background, foreground string
		profile                      colorprofile.Profile
	}{
		{"dark", "#10141c", "#e8e8e8", colorprofile.TrueColor},
		{"light", "#fafafa", "#202020", colorprofile.TrueColor},
		{"mono", "#10141c", "#e8e8e8", colorprofile.ASCII},
	} {
		t.Run(theme.name, func(t *testing.T) {
			e := testEnv(newFake())
			e.Styles.setProfile(theme.profile)
			e.Styles.background, e.Styles.foreground = lipgloss.Color(theme.background), lipgloss.Color(theme.foreground)
			var projects []tasksctl.Project
			for i, prefix := range []string{"alpha", "bravo", "charlie", "delta"} {
				rgb := []string{"#7995cd", "#0ba880", "#d2647a", "#bc8369"}[i]
				index := 42 + i
				stamp := now.Add(-[]time.Duration{0, time.Hour, 24 * time.Hour, 30 * 24 * time.Hour}[i]).Format(time.RFC3339)
				counts := []tasksctl.Counts{{Doing: 1, Todo: 3, Idea: 2}, {Doing: 4, Todo: 40, Idea: 8, Blocked: 1},
					{Doing: 12, Todo: 239, Idea: 20, Blocked: 2}, {Todo: 1000000}}[i]
				projects = append(projects, tasksctl.Project{Prefix: prefix, Root: "/demo/" + prefix, Reachable: true, Color: &rgb, Color256: &index, Counts: &counts, LastActivity: &stamp})
			}
			home := newProjectsView(e)
			home.main.Request(func(uint64) tea.Cmd { return nil })
			home.update(loadMsg{loader: home.main.ID(), gen: home.main.gen, data: projectsData{doing: 17, projects: tasksctl.ProjectsResult{Projects: projects}}})
			home.sel = 1
			app := New(e, Options{Stack: []view{home}})
			app.width, app.height = 80, 12
			captures := map[string]string{"home": home.renderAt(80, 10, now) + "\n" + app.statusLine()}
			rows := []rowView{
				fromRow(tasksctl.Row{ID: "bravo-aaa111", Title: "Run the next check", Status: "doing", Priority: 1, Updated: now.Format(time.RFC3339)}),
				fromRow(tasksctl.Row{ID: "bravo-bbb222", Title: "Review a result", Status: "todo", Priority: 2, Updated: now.Add(-time.Hour).Format(time.RFC3339)}),
				fromRow(tasksctl.Row{ID: "bravo-ccc333", Title: "Explore a possibility", Status: "idea", Priority: 3, Updated: now.Add(-24 * time.Hour).Format(time.RFC3339)}),
				fromRow(tasksctl.Row{ID: "bravo-ddd444", Title: "A completed item", Status: "done", Priority: 4, Updated: now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)}),
			}
			project := newProjectView(e, "bravo")
			project.rows, project.all, project.counts = rows, rows, *projects[1].Counts
			project.hasData, project.hasCount[0], project.shown[0] = true, true, len(rows)
			captures["project"] = project.renderAt(100, 12, now)
			task := newTaskView(e, target{ID: "bravo-aaa111", Prefix: "bravo"})
			task.data = &taskData{res: tasksctl.ShowResult{Task: tasksctl.Task{ID: "bravo-aaa111", Title: "Run the next check", Status: "doing", Priority: 1,
				Created: now.Add(-24 * time.Hour).Format(time.RFC3339), Updated: now.Format(time.RFC3339), Body: "A **clear result**, *next step*, and `command` with a [reference](https://example.invalid)."}}}
			task.width = 80
			captures["task"] = task.content()
			picker := &pickerOverlay{styles: e.Styles, prefix: "bravo", title: "launch", items: []string{"claude", "codex", "crush"}, sel: 1}
			captures["picker"] = picker.render(80)
			if plain := ansi.Strip(captures["home"]); !strings.Contains(plain, "▸ bravo") || strings.Contains(plain, "/demo/") || !strings.Contains(plain, "1000000") {
				t.Fatalf("home preview lost the requested structure: %s", plain)
			}
			if theme.profile == colorprofile.ASCII && !strings.Contains(ansi.Strip(captures["picker"]), "[2 codex]") {
				t.Fatal("monochrome picker lost selection")
			}
			for name, capture := range captures {
				for _, line := range strings.Split(capture, "\n") {
					if lipgloss.Width(line) > 100 {
						t.Fatalf("%s capture overflow", name)
					}
				}
				if dir := os.Getenv("TUI_PREVIEW_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, theme.name+"-"+name+".ansi"), []byte(capture), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
