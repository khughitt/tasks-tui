package ui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tasks-tui/internal/quickadd"
	"tasks-tui/internal/tasksctl"
)

// styledLine renders the input with each token in its class colour (spec §6.2).
func styledLine(s *Styles, line string, tokens []quickadd.Token, slot int) string {
	var b strings.Builder
	pos := 0
	for _, t := range tokens {
		if t.Start < pos || t.End > len(line) {
			continue
		}
		b.WriteString(line[pos:t.Start])
		text := line[t.Start:t.End]
		switch t.Kind {
		case quickadd.KindTag, quickadd.KindProject:
			b.WriteString(s.Accent(slot).Render(text))
		case quickadd.KindPriority:
			b.WriteString(s.PriorityStyle(int(text[1] - '0')).Render(text))
		case quickadd.KindSize, quickadd.KindComplexity:
			b.WriteString(s.Muted.Render(text))
		case quickadd.KindEvery:
			b.WriteString(s.Periodic.Render(text))
		case quickadd.KindIdea, quickadd.KindSeparator:
			b.WriteString(s.Muted.Italic(true).Render(text))
		case quickadd.KindError:
			b.WriteString(s.Error.Render(text))
		default:
			b.WriteString(text)
		}
		pos = t.End
	}
	b.WriteString(line[pos:])
	return b.String()
}

func quoteArg(arg string) string {
	if strings.ContainsAny(arg, " \t\"") {
		return `"` + strings.ReplaceAll(arg, `"`, `\"`) + `"`
	}
	return arg
}

func quoteArgs(args []string) string {
	out := make([]string, len(args))
	for i, arg := range args {
		out[i] = quoteArg(arg)
	}
	return strings.Join(out, " ")
}

func styledArgs(s *Styles, spec quickadd.Spec, slot int) string {
	parts := []string{quoteArgs([]string{"add", spec.Title}), s.Accent(slot).Render(quoteArgs([]string{"--project", spec.Project}))}
	add := func(style lipgloss.Style, flag, value string) {
		parts = append(parts, style.Render(quoteArgs([]string{flag, value})))
	}
	if spec.Idea {
		add(s.Muted.Italic(true), "--status", "idea")
	}
	if spec.Priority != nil {
		add(s.PriorityStyle(*spec.Priority), "-p", strconv.Itoa(*spec.Priority))
	}
	if spec.Size != "" {
		add(s.Muted, "--size", spec.Size)
	}
	if spec.Complexity != "" {
		add(s.Muted, "--complexity", spec.Complexity)
	}
	if spec.Every != "" {
		add(s.Periodic, "--every", spec.Every)
	}
	for _, tag := range spec.Tags {
		add(s.Accent(slot), "--tag", tag)
	}
	if spec.Body != "" {
		parts = append(parts, s.Muted.Italic(true).Render("-b"), quoteArgs([]string{spec.Body}))
	}
	return strings.Join(parts, " ")
}

func (a *App) openQuickAdd() tea.Cmd {
	s := a.env.Styles
	defaultProject := a.top().project()
	p := newPrompt(s, "quick add", "title #tag !2 ~m ^mid @30d >prefix  (? first for an idea, -- body)")
	p.hint = func(value string) string {
		if strings.TrimSpace(value) == "" {
			return s.Muted.Render("→ into " + defaultProject)
		}
		spec, err := quickadd.Parse(value, a.env.Prefixes, defaultProject)
		styled := styledLine(s, value, spec.Tokens, a.env.slot(spec.Project))
		if err != nil {
			return styled + "\n" + s.Error.Render(err.Error())
		}
		return styled + "\n" + s.Muted.Render("→ ") + styledArgs(s, spec, a.env.slot(spec.Project))
	}
	p.submit = func(value string) (overlay, tea.Cmd) {
		spec, err := quickadd.Parse(value, a.env.Prefixes, defaultProject)
		if err != nil {
			p.problem = err.Error()
			return p, nil
		}
		return nil, func() tea.Msg {
			ctx, cancel := a.env.ctx()
			defer cancel()
			res, err := a.env.Client.Add(ctx, spec.Args())
			return addMsg{res: res, err: err}
		}
	}
	p.input.ShowSuggestions = true
	a.overlay = &quickAddOverlay{prompt: p, env: a.env, defaultProject: defaultProject, loader: NewLoader()}
	return nil
}

type tagData struct {
	prefix string
	res    tasksctl.TagsResult
}

type quickAddOverlay struct {
	prompt                 *promptOverlay
	env                    *Env
	defaultProject, prefix string
	loader                 *Loader
	tags                   []string
}

func (o *quickAddOverlay) update(msg tea.Msg) (overlay, tea.Cmd) {
	if msg, ok := msg.(loadMsg); ok {
		if msg.loader != o.loader.ID() {
			return o, nil
		}
		accept, next := o.loader.Done(msg.gen)
		d := msg.data.(tagData)
		if !accept || d.prefix != o.prefix {
			return o, next
		}
		command := "tags --project " + d.prefix
		if msg.err != nil {
			return o, tea.Batch(next, notices(LevelError, command+": "+msg.err.Error()))
		}
		o.tags = nil
		for _, tag := range d.res.Tags {
			o.tags = append(o.tags, tag.Tag)
		}
		return o, tea.Batch(next, o.suggest(), notices(LevelWarning, prefixed(command, d.res.Warnings)...))
	}
	next, cmd := o.prompt.update(msg)
	if next == nil {
		return nil, cmd
	}
	return o, tea.Batch(cmd, o.suggest())
}

func (o *quickAddOverlay) suggest() tea.Cmd {
	prefix, suggestions := quickadd.TagSuggestions(o.prompt.input.Value(), o.prompt.input.Position(), o.env.Prefixes, o.defaultProject, o.tags)
	if prefix == o.prefix {
		o.prompt.input.SetSuggestions(suggestions)
		return nil
	}
	o.prefix, o.tags = prefix, nil
	o.prompt.input.SetSuggestions(nil)
	if prefix == "" {
		return nil
	}
	return o.loader.Request(func(gen uint64) tea.Cmd {
		return o.loader.Cmd(gen, func() (any, error) {
			ctx, cancel := o.env.ctx()
			defer cancel()
			res, err := o.env.Client.Tags(ctx, prefix)
			return tagData{prefix: prefix, res: res}, err
		})
	})
}

func (o *quickAddOverlay) render(width int) string {
	content := o.prompt.render(width)
	if len(o.prompt.input.MatchedSuggestions()) > 0 {
		content += "\n" + o.env.Styles.Muted.Render("tab complete tag · ↑/↓ choose")
	}
	return content
}
