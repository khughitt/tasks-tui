package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/khughitt/tasks-tui/internal/tasksctl"
)

var parkReasons = []string{"", "review", "decision", "approval", "environment", "dependency", "session"}

var generateUUID = newUUIDv4

func newUUIDv4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func isHumanRequestPark(spec tasksctl.ParkSpec) bool {
	return spec.WaitingOnUser && (spec.Reason == "review" || spec.Reason == "decision" || spec.Reason == "approval")
}

func isLowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// parseArtifactPin parses either:
//   - commit pin: "<repo>:<relpath>@<40-or-64-lower-hex>"
//   - digest pin: "<relpath>@<64-lower-hex>"
func parseArtifactPin(raw string) (repo, path, commit, sha256 string, err error) {
	s := strings.TrimSpace(raw)
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return "", "", "", "", fmt.Errorf("use repo:path@commit (40/64 hex) or path@sha256 (64 hex)")
	}
	left, hash := strings.TrimSpace(s[:at]), strings.TrimSpace(s[at+1:])
	if !isLowerHex(hash) {
		return "", "", "", "", fmt.Errorf("pin hash must be lowercase hex")
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return "", "", "", "", fmt.Errorf("pin hash must be valid hex")
	}
	if colon := strings.Index(left, ":"); colon >= 0 {
		repo = strings.TrimSpace(left[:colon])
		path = strings.TrimSpace(left[colon+1:])
		if repo == "" || path == "" || strings.HasPrefix(path, "/") {
			return "", "", "", "", fmt.Errorf("commit pin requires non-empty repo and relative path")
		}
		if len(hash) != 40 && len(hash) != 64 {
			return "", "", "", "", fmt.Errorf("commit hash must be 40 or 64 lowercase hex chars")
		}
		return repo, path, hash, "", nil
	}
	path = left
	if path == "" || strings.HasPrefix(path, "/") {
		return "", "", "", "", fmt.Errorf("digest pin requires a relative artifact path")
	}
	if len(hash) != 64 {
		return "", "", "", "", fmt.Errorf("sha256 digest must be 64 lowercase hex chars")
	}
	return "", path, "", hash, nil
}

func (a *App) selected() *target {
	t := a.top().current()
	if t == nil {
		a.log.Add(LevelInfo, "no task highlighted")
	}
	return t
}

func (a *App) write(action string, tgt target, force bool, run func(dir string) (tasksctl.WriteResult, error)) tea.Cmd {
	co := a.env.checkout(tgt)
	if co.Notice != "" {
		a.log.Add(LevelWarning, co.Notice)
	}
	return func() tea.Msg {
		res, err := run(co.Dir)
		return writeMsg{action: action, tgt: tgt, force: force, res: res, err: err}
	}
}

func (a *App) startTarget(force bool) tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	return a.startWith(*t, force)
}

func (a *App) startWith(tgt target, force bool) tea.Cmd {
	return a.write("start", tgt, force, func(dir string) (tasksctl.WriteResult, error) {
		ctx, cancel := a.env.ctx()
		defer cancel()
		return a.env.Client.Start(ctx, dir, tgt.ID, force)
	})
}

func (a *App) openForce(tgt target, detail string) tea.Cmd {
	a.overlay = &confirmOverlay{styles: a.env.Styles, text: a.env.Styles.Warning.Render(detail), accept: "F", label: "start --force",
		onAccept: func() tea.Cmd { return a.startWith(tgt, true) }}
	return nil
}

func (a *App) openPark() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	s := a.env.Styles
	spec := tasksctl.ParkSpec{}
	reason := 0
	p := newPrompt(s, "park "+tgt.ID+" — next step", "the one concrete next step")
	p.required = "next step is required"
	p.keys = func(msg tea.KeyPressMsg) bool {
		switch msg.String() {
		case "ctrl+u":
			spec.WaitingOnUser = !spec.WaitingOnUser
			return true
		case "ctrl+r":
			reason = (reason + 1) % len(parkReasons)
			spec.Reason = parkReasons[reason]
			return true
		}
		return false
	}
	p.hint = func(string) string {
		who := "agent"
		if spec.WaitingOnUser {
			who = "user"
		}
		r := spec.Reason
		if r == "" {
			r = "none"
		}
		return s.Base.Render("ctrl+u ") + "waiting on: " + who + s.Base.Render("   ctrl+r ") + "reason: " + r
	}
	p.submit = func(value string) (overlay, tea.Cmd) {
		spec.NextStep = strings.TrimSpace(value)
		if !isHumanRequestPark(spec) {
			return nil, a.submitPark(tgt, spec)
		}
		if tv, ok := a.top().(*taskView); ok && tv.data != nil && tv.tgt.ID == tgt.ID {
			return a.parkDecisionOverlay(tgt, spec, tv.data.res.Decisions), nil
		}
		co := a.env.checkout(tgt)
		if co.Notice != "" {
			a.log.Add(LevelWarning, co.Notice)
		}
		return nil, func() tea.Msg {
			ctx, cancel := a.env.ctx()
			defer cancel()
			res, err := a.env.Client.Show(ctx, co.Dir, tgt.ID)
			return parkShowMsg{tgt: tgt, spec: spec, res: res, err: err}
		}
	}
	a.overlay = p
	return nil
}

func (a *App) submitPark(tgt target, spec tasksctl.ParkSpec) tea.Cmd {
	specCopy := spec
	co := a.env.checkout(tgt)
	if co.Notice != "" {
		a.log.Add(LevelWarning, co.Notice)
	}
	return func() tea.Msg {
		ctx, cancel := a.env.ctx()
		defer cancel()
		res, err := a.env.Client.Park(ctx, co.Dir, tgt.ID, specCopy)
		return writeMsg{action: "park", tgt: tgt, parkSpec: &specCopy, res: res, err: err}
	}
}

func (a *App) openParkRetry(tgt target, spec tasksctl.ParkSpec, detail string) tea.Cmd {
	a.log.Add(LevelError, "park "+tgt.ID+": "+detail)
	s := a.env.Styles
	a.overlay = &confirmOverlay{
		styles: s,
		text:   s.Warning.Render("park failed (decision " + spec.DecisionID[:8] + " retained): " + detail),
		accept: "r",
		label:  "retry park",
		onAccept: func() tea.Cmd {
			return a.submitPark(tgt, spec)
		},
	}
	return nil
}

func (a *App) afterParkShow(msg parkShowMsg) tea.Cmd {
	if msg.err != nil {
		a.log.Add(LevelError, "park "+msg.tgt.ID+": show: "+msg.err.Error())
		return nil
	}
	a.overlay = a.parkDecisionOverlay(msg.tgt, msg.spec, msg.res.Decisions)
	return nil
}

func (a *App) parkDecisionOverlay(tgt target, spec tasksctl.ParkSpec, decisions []tasksctl.LogicalDecision) overlay {
	var matching []tasksctl.LogicalDecision
	for _, d := range decisions {
		if d.State != "open" {
			continue
		}
		if spec.Reason == "approval" && d.Data.Kind != "approval" {
			continue
		}
		matching = append(matching, d)
	}
	if len(matching) == 0 {
		return a.inlineParkDecisionPrompt(tgt, spec)
	}
	items := make([]string, 0, len(matching)+1)
	byLabel := make(map[string]string, len(matching))
	for _, d := range matching {
		shortID := d.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		text := strings.TrimSpace(d.Data.Text)
		if len(text) > 36 {
			text = text[:36] + "…"
		}
		label := fmt.Sprintf("%s %s: %s", d.Data.Kind, shortID, text)
		items = append(items, label)
		byLabel[label] = d.ID
	}
	const createNew = "new inline decision"
	items = append(items, createNew)
	return &pickerOverlay{
		styles: a.env.Styles,
		title:  "park " + tgt.ID + " — link open decision",
		items:  items,
		prefix: tgt.Prefix,
		onPick: func(label string) tea.Cmd {
			if label == createNew {
				next := a.inlineParkDecisionPrompt(tgt, spec)
				return func() tea.Msg { return setOverlayMsg{o: next} }
			}
			spec.DecisionID = byLabel[label]
			return a.submitPark(tgt, spec)
		},
	}
}

func (a *App) inlineParkDecisionPrompt(tgt target, spec tasksctl.ParkSpec) overlay {
	s := a.env.Styles
	spec.DecisionID = generateUUID()
	spec.DecisionOrigin = "tasks-tui"
	if spec.Reason == "approval" {
		spec.DecisionKind = "approval"
		p := newPrompt(s, "park "+tgt.ID+" — approval request", "what needs approval")
		p.required = "approval text is required"
		p.submit = func(value string) (overlay, tea.Cmd) {
			spec.DecisionText = strings.TrimSpace(value)
			pinPrompt := newPrompt(s, "park "+tgt.ID+" — artifact pin", "repo:path@commit or path@sha256")
			pinPrompt.required = "artifact pin is required for approval"
			pinPrompt.hint = func(string) string {
				return s.Base.Render("pin: ") + "repo:rel/path@<40|64 hex> or rel/path@<64 hex>"
			}
			pinPrompt.submit = func(pinRaw string) (overlay, tea.Cmd) {
				repo, path, commit, sha, err := parseArtifactPin(pinRaw)
				if err != nil {
					pinPrompt.problem = err.Error()
					return pinPrompt, nil
				}
				spec.ArtifactRepo = repo
				spec.ArtifactPath = path
				spec.ArtifactCommit = commit
				spec.ArtifactSHA256 = sha
				return nil, a.submitPark(tgt, spec)
			}
			return pinPrompt, nil
		}
		return p
	}
	spec.DecisionKind = "question"
	p := newPrompt(s, "park "+tgt.ID+" — decision question", "what the user needs to decide or review")
	p.required = "decision text is required"
	p.submit = func(value string) (overlay, tea.Cmd) {
		spec.DecisionText = strings.TrimSpace(value)
		return nil, a.submitPark(tgt, spec)
	}
	return p
}

func (a *App) openDone() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	p := newPrompt(a.env.Styles, "done "+tgt.ID+" — what landed", "optional message")
	p.submit = func(value string) (overlay, tea.Cmd) {
		message := strings.TrimSpace(value)
		return nil, a.write("done", tgt, false, func(dir string) (tasksctl.WriteResult, error) {
			ctx, cancel := a.env.ctx()
			defer cancel()
			return a.env.Client.Done(ctx, dir, tgt.ID, message)
		})
	}
	a.overlay = p
	return nil
}

func (a *App) openDrop() tea.Cmd {
	t := a.selected()
	if t == nil {
		return nil
	}
	tgt := *t
	s := a.env.Styles
	p := newPrompt(s, "drop "+tgt.ID+" — why", "optional message")
	p.submit = func(value string) (overlay, tea.Cmd) {
		message := strings.TrimSpace(value)
		return &confirmOverlay{styles: s, text: "really drop " + tgt.ID + "? " + s.Base.Render(strings.TrimSpace(tgt.Title)), accept: "y", label: "drop",
			onAccept: func() tea.Cmd {
				return a.write("drop", tgt, false, func(dir string) (tasksctl.WriteResult, error) {
					ctx, cancel := a.env.ctx()
					defer cancel()
					return a.env.Client.Drop(ctx, dir, tgt.ID, message)
				})
			}}, nil
	}
	a.overlay = p
	return nil
}
