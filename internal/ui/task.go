package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

type taskData struct {
	res      tasksctl.ShowResult
	checkout tasksctl.Checkout
}

type taskView struct {
	env    *Env
	tgt    target
	loader *Loader
	data   *taskData
	vp     viewport.Model
	width  int
}

func newTaskView(env *Env, t target) *taskView {
	return &taskView{env: env, tgt: t, loader: NewLoader(), vp: viewport.New()}
}
func (v *taskView) title() string   { return v.tgt.ID }
func (v *taskView) project() string { return v.tgt.Prefix }
func (v *taskView) capturing() bool { return false }
func (v *taskView) loading() bool   { return v.loader.InFlight() }

// warnings is the checkout notice, then show's, once the first load has landed.
func (v *taskView) warnings() []string {
	if v.data == nil {
		return nil
	}
	var ws []string
	if v.data.checkout.Notice != "" {
		ws = append(ws, v.data.checkout.Notice)
	}
	return append(ws, prefixed("show "+v.tgt.ID, v.data.res.Warnings)...)
}
func (v *taskView) current() *target {
	t := v.tgt
	if v.data != nil {
		t.Title, t.Park, t.Claim = v.data.res.Task.Title, v.data.res.Park, v.data.res.Claim
	}
	return &t
}

func (v *taskView) reload() tea.Cmd {
	tgt := *v.current()
	return v.loader.Request(func(gen uint64) tea.Cmd {
		return v.loader.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			co := v.env.checkout(tgt)
			res, err := v.env.Client.Show(ctx, co.Dir, tgt.ID)
			if err != nil {
				return nil, err
			}
			return taskData{res: res, checkout: co}, nil
		})
	})
}

func (v *taskView) update(msg tea.Msg) (view, tea.Cmd) {
	switch msg := msg.(type) {
	case loadMsg:
		if msg.loader != v.loader.ID() {
			return v, nil
		}
		accept, next := v.loader.Done(msg.gen)
		if !accept || msg.background {
			return v, next
		}
		if msg.err != nil {
			return v, tea.Batch(next, notices(LevelError, "show "+v.tgt.ID+": "+msg.err.Error()))
		}
		d := msg.data.(taskData)
		v.data = &d
		v.tgt.Park, v.tgt.Claim = d.res.Park, d.res.Claim
		v.vp.SetContent(v.content())
		return v, next
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Down):
			v.vp.ScrollDown(1)
		case key.Matches(msg, keys.Up):
			v.vp.ScrollUp(1)
		case key.Matches(msg, keys.PageDown):
			v.vp.PageDown()
		case key.Matches(msg, keys.PageUp):
			v.vp.PageUp()
		case key.Matches(msg, keys.Top):
			v.vp.GotoTop()
		case key.Matches(msg, keys.Copy):
			return v, tea.SetClipboard(v.tgt.ID)
		}
	}
	return v, nil
}

func (v *taskView) render(width, height int) string {
	if v.width != width {
		v.width = width
		if v.data != nil {
			v.vp.SetContent(v.content())
		}
	}
	v.vp.SetWidth(width)
	v.vp.SetHeight(height)
	if v.data == nil {
		return v.env.Styles.Muted.Render("loading " + v.tgt.ID + " …")
	}
	return v.vp.View()
}

func (v *taskView) status(status string, slot int) lipgloss.Style {
	style := v.env.Styles.StatusStyle(status)
	if status == "doing" {
		style = style.Foreground(v.env.Styles.Accent(slot).GetForeground())
	}
	return style
}

func (v *taskView) content() string {
	s, r := v.env.Styles, v.data.res
	t, slot := r.Task, v.env.slot(r.Task.Prefix())
	var b strings.Builder
	line := func(k, val string) {
		if val != "" {
			fmt.Fprintf(&b, "%s %s\n", s.Muted.Render(pad(k, 11)), val)
		}
	}
	fmt.Fprintf(&b, "%s %s\n", s.Accent(slot).Render(t.ID), s.Header.Render(t.Title))
	line("status", v.status(t.Status, slot).Render(t.Status))
	line("priority", s.PriorityStyle(t.Priority).Render(fmt.Sprintf("P%d", t.Priority)))
	line("size", deref(t.Size))
	line("complexity", deref(t.Complexity))
	line("process", deref(t.Process))
	line("owner", deref(t.Owner))
	line("created", t.Created)
	line("started", deref(t.Started))
	line("updated", t.Updated)
	line("completed", deref(t.Completed))
	line("tags", strings.Join(t.Tags, ", "))
	line("source", deref(t.Source))
	if r.Periodic != nil {
		due := "not due"
		if r.Periodic.Due != nil {
			due = "due " + *r.Periodic.Due
		}
		line("every", s.Periodic.Render(r.Periodic.Every+"  "+due))
	}
	if t.Spec != nil {
		line("spec", *t.Spec+"  "+s.Muted.Render(deref(r.SpecPath)))
	}
	if t.Plan != nil {
		step := deref(t.Step)
		if r.StepFound != nil && !*r.StepFound {
			step += s.Error.Render("  (step not found)")
		}
		line("plan", *t.Plan+"  "+s.Muted.Render(deref(r.PlanPath))+"\n            "+step)
	}
	line("checkout", s.Muted.Render(v.data.checkout.Dir))
	if c := r.Claim; c != nil {
		state := s.Claim.Foreground(s.Accent(slot).GetForeground()).Render("live")
		if !c.Live {
			state = s.Stale.Render("stale")
		}
		line("claim", fmt.Sprintf("%s %s  %s  %s", state, c.Owner, c.Session, s.Muted.Render(c.Worktree)))
	}
	if p := r.Park; p != nil {
		waiting := s.ParkAgent.Render("waiting on agent")
		if p.WaitingOn == "user" {
			waiting = s.ParkUser.Render("waiting on user")
		}
		reason := ""
		if p.Reason != nil {
			reason = "  " + *p.Reason
		}
		line("parked", waiting+reason+"  "+s.Muted.Render(p.Worktree)+"\n            → "+p.NextStep)
	}
	if e := r.Escalation; e != nil {
		line("escalated", s.Warning.Render(e.Level)+"  "+s.Muted.Render(e.At))
	}
	b.WriteString("\n")
	if strings.TrimSpace(t.Body) != "" {
		body, err := v.markdown(t.Body)
		if err != nil {
			fmt.Fprintf(&b, "%s %v\n\n%s", s.Error.Render("body render failed:"), err, t.Body)
		} else {
			b.WriteString(body)
		}
		b.WriteString("\n")
	}
	if len(t.Notes) > 0 {
		b.WriteString(s.Header.Render("notes") + "\n")
		for _, n := range t.Notes {
			fmt.Fprintf(&b, "%s %s\n", s.Muted.Render(n.At+" ("+n.By+")"), n.Text)
		}
		b.WriteString("\n")
	}
	relations := func(title string, rows []tasksctl.Relation) {
		if len(rows) == 0 {
			return
		}
		b.WriteString(s.Header.Render(title) + "\n")
		for _, x := range rows {
			xslot := v.env.slot(prefixOf(x.ID))
			fmt.Fprintf(&b, "  %s %s %s\n", s.Accent(xslot).Render(x.ID), v.status(x.Status, xslot).Render(pad(x.Status, 7)), x.Title)
		}
	}
	if r.Parent != nil {
		relations("parent", []tasksctl.Relation{*r.Parent})
	}
	relations("children", r.Children)
	if len(r.DependsOn) > 0 {
		b.WriteString(s.Header.Render("depends on") + "\n")
		for _, x := range r.DependsOn {
			status, title := "unresolved", s.Muted.Render("(not reachable from here)")
			if x.Resolved {
				status, title = deref(x.Status), deref(x.Title)
			}
			xslot := v.env.slot(prefixOf(x.ID))
			fmt.Fprintf(&b, "  %s %s %s\n", s.Accent(xslot).Render(x.ID), v.status(status, xslot).Render(pad(status, 10)), title)
		}
	}
	return b.String()
}

func (v *taskView) markdown(src string) (string, error) {
	width := v.width
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(v.env.Styles.GlamourStyle()), glamour.WithWordWrap(width-2))
	if err != nil {
		return "", err
	}
	return r.Render(src)
}

func prefixOf(id string) string {
	if i := strings.LastIndex(id, "-"); i >= 0 {
		return id[:i]
	}
	return id
}
