package ui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

func openDecisions(all []tasksctl.LogicalDecision) []tasksctl.LogicalDecision {
	var out []tasksctl.LogicalDecision
	for _, d := range all {
		if d.State == "open" {
			out = append(out, d)
		}
	}
	return out
}

func shortUUID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatAgeSeconds(sec *int64) string {
	if sec == nil {
		return "age unknown"
	}
	s := *sec
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	if s < 3600 {
		return fmt.Sprintf("%dm", s/60)
	}
	if s < 86400 {
		return fmt.Sprintf("%dh", s/3600)
	}
	return fmt.Sprintf("%dd", s/86400)
}

func formatDecisionSubject(subj tasksctl.DecisionSubject, changed *bool) string {
	switch {
	case subj.Commit != nil && subj.Path != nil && subj.Repo != nil:
		return fmt.Sprintf("pin %s:%s@%s", *subj.Repo, *subj.Path, shortUUID(*subj.Commit))
	case subj.SHA256 != nil && subj.Path != nil:
		return fmt.Sprintf("pin %s@sha256:%s", *subj.Path, shortUUID(*subj.SHA256))
	case subj.Scope != nil:
		if changed != nil && *changed {
			return "scope CHANGED since open"
		}
		return "scope intact"
	default:
		return "scope unknown"
	}
}

func formatDependentsBadge(deps tasksctl.OpenDependents) string {
	var parts []string
	if len(deps.Tasks) > 0 {
		ids := make([]string, 0, len(deps.Tasks))
		for _, t := range deps.Tasks {
			label := t.ID
			if t.Title != "" {
				label += " (" + t.Title + ")"
			}
			ids = append(ids, label)
		}
		parts = append(parts, fmt.Sprintf("%d open dep(s): %s", len(deps.Tasks), strings.Join(ids, ", ")))
	}
	if !deps.Complete {
		parts = append(parts, "coverage incomplete: "+strings.Join(deps.UnavailableProjects, ", "))
	}
	return strings.Join(parts, " · ")
}

func renderDecisionsSection(b *strings.Builder, s *Styles, decisions []tasksctl.LogicalDecision) {
	b.WriteString(s.Header.Render("decisions") + "\n")
	hasOpen := false
	for _, d := range decisions {
		if d.State == "open" {
			hasOpen = true
		}
		stateStyle := s.Base
		if d.State == "open" {
			stateStyle = s.Warning
		}
		fmt.Fprintf(b, "  [%s] %s %s (%s by %s at %s)\n",
			stateStyle.Render(d.State), d.Data.Kind, shortUUID(d.ID), d.Data.Origin, d.By, d.At)
		fmt.Fprintf(b, "    %s\n", d.Data.Text)
		subjLine := formatDecisionSubject(d.Subject, d.ScopeChangedSinceOpen)
		if d.ScopeChangedSinceOpen != nil && *d.ScopeChangedSinceOpen {
			subjLine = s.Warning.Render(subjLine)
		} else {
			subjLine = s.Base.Render(subjLine)
		}
		fmt.Fprintf(b, "    %s", subjLine)
		if deps := formatDependentsBadge(d.OpenDependents); deps != "" {
			fmt.Fprintf(b, " · %s", s.Warning.Render(deps))
		}
		b.WriteByte('\n')
		if d.Resolution != nil {
			fmt.Fprintf(b, "    resolved: %s by %s at %s — %s\n",
				d.Resolution.Data.Outcome, d.Resolution.By, d.Resolution.At, d.Resolution.Data.Text)
		}
	}
	if hasOpen {
		b.WriteString("  " + s.Base.Render("enter: answer or withdraw an open decision") + "\n")
	}
	b.WriteByte('\n')
}

func (a *App) decisionSelectorOverlay(tgt target, decisions []tasksctl.LogicalDecision) overlay {
	if len(decisions) == 0 {
		return nil
	}
	if len(decisions) == 1 {
		return a.decisionActionOverlay(tgt, decisions[0])
	}
	items := make([]string, 0, len(decisions))
	byLabel := make(map[string]tasksctl.LogicalDecision, len(decisions))
	for _, d := range decisions {
		text := strings.TrimSpace(d.Data.Text)
		if len(text) > 32 {
			text = text[:32] + "…"
		}
		tag := ""
		if d.ScopeChangedSinceOpen != nil && *d.ScopeChangedSinceOpen {
			tag = " [scope changed]"
		}
		label := fmt.Sprintf("%s %s%s: %s", d.Data.Kind, shortUUID(d.ID), tag, text)
		items = append(items, label)
		byLabel[label] = d
	}
	return &pickerOverlay{
		styles: a.env.Styles,
		title:  "select decision on " + tgt.ID,
		items:  items,
		prefix: tgt.Prefix,
		onPick: func(label string) tea.Cmd {
			next := a.decisionActionOverlay(tgt, byLabel[label])
			return func() tea.Msg { return setOverlayMsg{o: next} }
		},
	}
}

func (a *App) decisionActionOverlay(tgt target, d tasksctl.LogicalDecision) overlay {
	var items []string
	var acceptLabel, declineLabel, replyLabel, withdrawLabel string
	switch d.Data.Kind {
	case "drop":
		token := ""
		if d.SeenScope != nil {
			token = shortUUID(*d.SeenScope)
		}
		acceptLabel = fmt.Sprintf("accept (DROPS %s atomically; scope %s)", tgt.ID, token)
		declineLabel = "decline (keeps task open)"
		withdrawLabel = "withdraw (cancels decision)"
		items = []string{acceptLabel, declineLabel, withdrawLabel}
	case "question":
		replyLabel = "reply (records answer)"
		withdrawLabel = "withdraw (cancels decision)"
		items = []string{replyLabel, withdrawLabel}
	default: // suggestion, approval
		acceptLabel = "accept (resolves decision; status unchanged)"
		declineLabel = "decline (resolves decision; status unchanged)"
		withdrawLabel = "withdraw (cancels decision)"
		items = []string{acceptLabel, declineLabel, withdrawLabel}
	}
	title := fmt.Sprintf("%s %s on %s — %s", d.Data.Kind, shortUUID(d.ID), tgt.ID, formatDecisionSubject(d.Subject, d.ScopeChangedSinceOpen))
	if deps := formatDependentsBadge(d.OpenDependents); deps != "" {
		title += " · " + deps
	}
	return &pickerOverlay{
		styles: a.env.Styles,
		title:  title,
		items:  items,
		prefix: tgt.Prefix,
		onPick: func(choice string) tea.Cmd {
			var outcome string
			switch choice {
			case acceptLabel:
				outcome = "accept"
			case declineLabel:
				outcome = "decline"
			case replyLabel:
				outcome = "reply"
			case withdrawLabel:
				outcome = "withdraw"
			default:
				return nil
			}
			next := a.decisionTextPrompt(tgt, d, outcome)
			return func() tea.Msg { return setOverlayMsg{o: next} }
		},
	}
}

func (a *App) decisionTextPrompt(tgt target, d tasksctl.LogicalDecision, outcome string) overlay {
	s := a.env.Styles
	title := fmt.Sprintf("%s %s %s on %s", outcome, d.Data.Kind, shortUUID(d.ID), tgt.ID)
	placeholder := "answer text"
	if outcome == "withdraw" {
		placeholder = "reason for withdrawal"
	}
	p := newPrompt(s, title, placeholder)
	p.required = "text is required"
	p.hint = func(string) string {
		if d.Data.Kind == "drop" && outcome == "accept" {
			tok := ""
			if d.SeenScope != nil {
				tok = shortUUID(*d.SeenScope)
			}
			return s.Warning.Render("effect: atomically drops "+tgt.ID) + s.Base.Render(" (seen_scope "+tok+")")
		}
		return s.Base.Render("decision: ") + d.Data.Text
	}
	p.submit = func(value string) (overlay, tea.Cmd) {
		text := strings.TrimSpace(value)
		co := a.env.checkout(tgt)
		if co.Notice != "" {
			a.log.Add(LevelWarning, co.Notice)
		}
		if outcome == "withdraw" {
			return nil, func() tea.Msg {
				ctx, cancel := a.env.ctx()
				defer cancel()
				res, err := a.env.Client.DecisionWithdraw(ctx, co.Dir, tgt.ID, d.ID, text)
				return decisionWriteMsg{action: "decision withdraw", tgt: tgt, decisionID: d.ID, res: res, err: err}
			}
		}
		spec := tasksctl.DecisionAnswerSpec{Outcome: outcome, Text: text}
		if d.Data.Kind == "drop" && outcome == "accept" {
			if d.SeenScope == nil || *d.SeenScope == "" {
				p.problem = "missing displayed seen_scope token; refresh task first"
				return p, nil
			}
			spec.SeenScope = *d.SeenScope
		}
		return nil, func() tea.Msg {
			ctx, cancel := a.env.ctx()
			defer cancel()
			res, err := a.env.Client.DecisionAnswer(ctx, co.Dir, tgt.ID, d.ID, spec)
			return decisionWriteMsg{action: "decision answer", tgt: tgt, decisionID: d.ID, res: res, err: err}
		}
	}
	return p
}

func (a *App) afterDecisionWrite(msg decisionWriteMsg) tea.Cmd {
	if msg.err != nil {
		var taskErr *tasksctl.Error
		if errors.As(msg.err, &taskErr) && taskErr.Kind == "stale_scope" {
			a.log.Add(LevelWarning, fmt.Sprintf("decision %s on %s: scope changed since displayed; refreshed for a new explicit answer (%s)",
				shortUUID(msg.decisionID), msg.tgt.ID, taskErr.Detail))
			return a.top().reload()
		}
		a.log.Add(LevelError, msg.action+" "+msg.tgt.ID+": "+msg.err.Error())
		return nil
	}
	a.log.AddWarnings(msg.action+" "+msg.tgt.ID, msg.res.Warnings)
	a.log.Add(LevelInfo, fmt.Sprintf("%s %s (%s)", msg.action, msg.tgt.ID, shortUUID(msg.decisionID)))
	return a.top().reload()
}

// pendingView lists pending human decisions and legacy waits from `tasks list --pending`.
// Separate decisions on the same task appear as separate rows. Legacy rows expose
// read-only evidence and open task detail rather than an answer overlay.
type pendingLoadData struct {
	res    tasksctl.PendingResult
	parks  map[string]*tasksctl.ParkInfo
	claims map[string]*tasksctl.ClaimInfo
}

type pendingView struct {
	env    *Env
	prefix string
	loader *Loader
	data   *tasksctl.PendingResult
	parks  map[string]*tasksctl.ParkInfo
	claims map[string]*tasksctl.ClaimInfo
	sel    int
	offset int
	warns  []string
}

func newPendingView(env *Env, prefix string) *pendingView {
	return &pendingView{env: env, prefix: prefix, loader: NewLoader()}
}

func (v *pendingView) title() string {
	if v.prefix != "" {
		return v.prefix + " pending"
	}
	return "pending"
}
func (v *pendingView) project() string    { return v.prefix }
func (v *pendingView) capturing() bool    { return false }
func (v *pendingView) loading() bool      { return v.loader.InFlight() }
func (v *pendingView) warnings() []string { return v.warns }

func (v *pendingView) selectedItem() *tasksctl.PendingItem {
	if v.data == nil || v.sel < 0 || v.sel >= len(v.data.Items) {
		return nil
	}
	return &v.data.Items[v.sel]
}

func (v *pendingView) targetFor(item *tasksctl.PendingItem) target {
	return target{
		ID:     item.Task.ID,
		Title:  item.Task.Title,
		Prefix: item.Task.Project,
		Park:   v.parks[item.Task.ID],
		Claim:  v.claims[item.Task.ID],
	}
}

func (v *pendingView) current() *target {
	item := v.selectedItem()
	if item == nil {
		return nil
	}
	t := v.targetFor(item)
	return &t
}

func (v *pendingView) reload() tea.Cmd {
	prefix := v.prefix
	return v.loader.Request(func(gen uint64) tea.Cmd {
		return v.loader.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			res, err := v.env.Client.Pending(ctx, prefix)
			if err != nil {
				return nil, err
			}
			parks := make(map[string]*tasksctl.ParkInfo)
			claims := make(map[string]*tasksctl.ClaimInfo)
			if parked, perr := v.env.Client.Parked(ctx, prefix); perr == nil {
				for _, r := range parked.Tasks {
					if r.Park != nil {
						parks[r.ID] = r.Park
					}
					if cl := r.LiveClaim(); cl != nil {
						claims[r.ID] = cl
					}
				}
			}
			if prefix != "" {
				if doing, derr := v.env.Client.List(ctx, prefix, tasksctl.ListOpts{Status: []string{"doing"}}); derr == nil {
					for _, r := range doing.Tasks {
						if r.Park != nil && parks[r.ID] == nil {
							parks[r.ID] = r.Park
						}
						if cl := r.LiveClaim(); cl != nil && claims[r.ID] == nil {
							claims[r.ID] = cl
						}
					}
				}
			}
			return pendingLoadData{res: res, parks: parks, claims: claims}, nil
		})
	})
}

func (v *pendingView) update(raw tea.Msg) (view, tea.Cmd) {
	switch msg := raw.(type) {
	case loadMsg:
		if msg.loader != v.loader.ID() {
			return v, nil
		}
		accept, next := v.loader.Done(msg.gen)
		if !accept || msg.background {
			return v, next
		}
		if msg.err != nil {
			return v, tea.Batch(next, notices(LevelError, "pending: "+msg.err.Error()))
		}
		loaded := msg.data.(pendingLoadData)
		res := loaded.res
		v.data = &res
		v.parks = loaded.parks
		v.claims = loaded.claims
		v.warns = prefixed("list --pending", res.Warnings)
		if !res.Coverage.Complete {
			v.warns = append(v.warns, "list --pending: incomplete coverage ("+strings.Join(res.Coverage.UnavailableProjects, ", ")+")")
		}
		if len(res.Items) == 0 {
			v.sel, v.offset = 0, 0
		} else if v.sel >= len(res.Items) {
			v.sel = len(res.Items) - 1
		}
		return v, next
	case tea.KeyPressMsg:
		if v.data == nil {
			return v, nil
		}
		switch {
		case key.Matches(msg, keys.Down):
			v.sel = min(v.sel+1, max(0, len(v.data.Items)-1))
		case key.Matches(msg, keys.Up):
			v.sel = max(0, v.sel-1)
		case key.Matches(msg, keys.Top):
			v.sel = 0
		case key.Matches(msg, keys.Bottom):
			v.sel = max(0, len(v.data.Items)-1)
		case key.Matches(msg, keys.Enter):
			item := v.selectedItem()
			if item == nil {
				return v, nil
			}
			tgt := v.targetFor(item)
			// Legacy rows never offer a direct answer overlay: open task detail for evidence.
			if item.Type != "decision" || item.Decision == nil {
				return v, func() tea.Msg { return pushMsg{v: newTaskView(v.env, tgt)} }
			}
			dec := *item.Decision
			return v, func() tea.Msg { return openDecisionMsg{tgt: tgt, decisions: []tasksctl.LogicalDecision{dec}} }
		}
	}
	return v, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatPendingRow(s *Styles, item tasksctl.PendingItem, selected bool) string {
	marker := "  "
	if selected {
		marker = "▸ "
	}
	age := formatAgeSeconds(item.AgeSeconds)
	if item.Type == "decision" && item.Decision != nil {
		d := item.Decision
		subj := formatDecisionSubject(d.Subject, d.ScopeChangedSinceOpen)
		deps := formatDependentsBadge(d.OpenDependents)
		extra := subj
		if deps != "" {
			extra += " · " + deps
		}
		return fmt.Sprintf("%s%s  %-10s  %s  %s  (%s)  %s — %s",
			marker, s.Accent(item.Task.Project).Render(item.Task.ID), d.Data.Kind, shortUUID(d.ID), age, extra, item.Task.Title, d.Data.Text)
	}
	ev := item.Source
	if item.Reason != nil && *item.Reason != "" {
		ev += "/" + *item.Reason
	}
	if item.Host != nil && *item.Host != "" {
		ev += "@" + *item.Host
	}
	text := ""
	if item.Text != nil {
		text = *item.Text
	}
	return fmt.Sprintf("%s%s  %-10s  %s  (legacy evidence only; no direct answer)  %s — %s",
		marker, s.Accent(item.Task.Project).Render(item.Task.ID), "legacy:"+ev, age, item.Task.Title, text)
}

func (v *pendingView) render(width, height int) string {
	s := v.env.Styles
	head := s.Header.Render("pending decisions")
	if v.data != nil && !v.data.Coverage.Complete {
		head += "  " + s.Warning.Render("incomplete coverage: "+strings.Join(v.data.Coverage.UnavailableProjects, ", "))
	}
	lines := []string{head}
	if v.data == nil {
		lines = append(lines, "  "+s.Muted.Render("loading pending …"))
		return lipgloss.NewStyle().MaxHeight(height).Render(strings.Join(lines, "\n"))
	}
	if len(v.data.Items) == 0 {
		lines = append(lines, "  "+s.Muted.Render("no pending human decisions"))
		return lipgloss.NewStyle().MaxHeight(height).Render(strings.Join(lines, "\n"))
	}
	clipped := lipgloss.NewStyle().MaxWidth(width)
	avail := max(1, height-len(lines))
	if v.sel < v.offset {
		v.offset = v.sel
	}
	if v.sel >= v.offset+avail {
		v.offset = v.sel - avail + 1
	}
	for i := v.offset; i < len(v.data.Items) && i < v.offset+avail; i++ {
		lines = append(lines, clipped.Render(formatPendingRow(s, v.data.Items[i], i == v.sel)))
	}
	return lipgloss.NewStyle().MaxHeight(height).Render(strings.Join(lines, "\n"))
}
