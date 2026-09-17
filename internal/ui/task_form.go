package ui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tasks-tui/internal/tasksctl"
)

type taskFormData struct {
	task        tasksctl.Task
	checkout    tasksctl.Checkout
	tags        []string
	warnings    []string
	tagWarnings []string
	tagErr      error
}

type formSubmitMsg struct {
	form     *taskFormView
	action   string
	warnings []string
	err      error
}

type formCloseMsg struct{}

type taskFormView struct {
	env         *Env
	tgt         target
	create      bool
	loader      *Loader
	loadStarted bool
	titleInput  textinput.Model
	body        textarea.Model
	tags        textinput.Model
	focus       int
	problem     string
	data        *taskFormData
	knownTags   []string
	submitting  bool
}

func newTaskFormView(env *Env, tgt target) *taskFormView {
	title := textinput.New()
	title.Prompt = ""
	titleStyles := textinput.DefaultStyles(env.Styles.Tone.Dark)
	titleStyles.Cursor.Blink = false
	title.SetStyles(titleStyles)

	body := textarea.New()
	body.Prompt = ""
	body.ShowLineNumbers = false
	body.SetHeight(6)
	bodyStyles := textarea.DefaultStyles(env.Styles.Tone.Dark)
	bodyStyles.Cursor.Blink = false
	body.SetStyles(bodyStyles)

	tags := textinput.New()
	tags.Prompt = ""
	tagsStyles := textinput.DefaultStyles(env.Styles.Tone.Dark)
	tagsStyles.Cursor.Blink = false
	tags.SetStyles(tagsStyles)
	tags.ShowSuggestions = true
	tags.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("enter"))

	v := &taskFormView{env: env, tgt: tgt, loader: NewLoader(), titleInput: title, body: body, tags: tags}
	v.focusField(0)
	return v
}

func newAddFormView(env *Env, prefix string) *taskFormView {
	v := newTaskFormView(env, target{Prefix: prefix})
	v.create = true
	v.data = &taskFormData{}
	return v
}

func (v *taskFormView) titleText() string {
	if v.create {
		return "add task"
	}
	return "edit " + v.tgt.ID
}
func (v *taskFormView) title() string    { return v.titleText() }
func (v *taskFormView) project() string  { return v.tgt.Prefix }
func (v *taskFormView) capturing() bool  { return true }
func (v *taskFormView) loading() bool    { return v.loader.InFlight() }
func (v *taskFormView) current() *target { return &v.tgt }

func (v *taskFormView) reload() tea.Cmd {
	if v.loadStarted {
		return nil
	}
	v.loadStarted = true
	co := v.env.checkout(v.tgt)
	return v.loader.Request(func(gen uint64) tea.Cmd {
		return v.loader.Cmd(gen, func() (any, error) {
			ctx, cancel := v.env.ctx()
			defer cancel()
			d := taskFormData{checkout: co}
			if co.Notice != "" {
				d.warnings = append(d.warnings, co.Notice)
			}
			if !v.create {
				res, err := v.env.Client.Show(ctx, co.Dir, v.tgt.ID)
				if err != nil {
					return nil, err
				}
				d.task = res.Task
				d.warnings = append(d.warnings, prefixed("show "+v.tgt.ID, res.Warnings)...)
			}
			tags, err := v.env.Client.Tags(ctx, v.tgt.Prefix)
			if err != nil {
				d.tagErr = err
			} else {
				d.tagWarnings = tags.Warnings
				for _, tag := range tags.Tags {
					d.tags = append(d.tags, tag.Tag)
				}
			}
			return d, nil
		})
	})
}

func (v *taskFormView) focusField(field int) {
	v.focus = field
	v.titleInput.Blur()
	v.body.Blur()
	v.tags.Blur()
	switch field {
	case 0:
		v.titleInput.Focus()
	case 1:
		v.body.Focus()
	default:
		v.tags.Focus()
	}
}

func (v *taskFormView) update(msg tea.Msg) (view, tea.Cmd) {
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
			v.problem = v.titleText() + ": " + msg.err.Error()
			return v, next
		}
		d := msg.data.(taskFormData)
		v.data = &d
		v.knownTags = d.tags
		if !v.create {
			v.titleInput.SetValue(d.task.Title)
			v.body.SetValue(d.task.Body)
			v.tags.SetValue(strings.Join(d.task.Tags, " "))
		}
		v.suggestTags()
		cmds := []tea.Cmd{next, notices(LevelWarning, d.warnings...)}
		if d.tagErr != nil {
			cmds = append(cmds, notices(LevelError, "tags --project "+v.tgt.Prefix+": "+d.tagErr.Error()))
		} else {
			cmds = append(cmds, notices(LevelWarning, prefixed("tags --project "+v.tgt.Prefix, d.tagWarnings)...))
		}
		return v, tea.Batch(cmds...)
	case formSubmitMsg:
		v.submitting = false
		v.problem = msg.action + ": " + msg.err.Error()
		return v, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if v.submitting {
				return v, nil
			}
			return v, func() tea.Msg { return formCloseMsg{} }
		case "tab":
			v.focusField((v.focus + 1) % 3)
			return v, nil
		case "shift+tab":
			v.focusField((v.focus + 2) % 3)
			return v, nil
		case "ctrl+s":
			return v, v.submit()
		}
		v.problem = ""
	}
	var cmd tea.Cmd
	switch v.focus {
	case 0:
		v.titleInput, cmd = v.titleInput.Update(msg)
	case 1:
		v.body, cmd = v.body.Update(msg)
	case 2:
		v.tags, cmd = v.tags.Update(msg)
		v.suggestTags()
	}
	return v, cmd
}

func (v *taskFormView) suggestTags() {
	value := v.tags.Value()
	start := strings.LastIndexAny(value, " \t") + 1
	partial := value[start:]
	used := strings.Fields(value[:start])
	var suggestions []string
	for _, tag := range v.knownTags {
		if strings.HasPrefix(tag, partial) && tag != partial && !slices.Contains(used, tag) {
			suggestions = append(suggestions, value[:start]+tag)
		}
	}
	v.tags.SetSuggestions(suggestions)
}

func (v *taskFormView) submit() tea.Cmd {
	if v.submitting {
		return nil
	}
	if v.data == nil {
		v.problem = "task is still loading"
		return nil
	}
	title := strings.TrimSpace(v.titleInput.Value())
	if title == "" {
		v.problem = "title is required"
		return nil
	}
	fields := tasksctl.TaskFields{Title: title, Body: v.body.Value(), Tags: strings.Fields(v.tags.Value())}
	v.submitting = true
	return func() tea.Msg {
		ctx, cancel := v.env.ctx()
		defer cancel()
		if v.create {
			res, err := v.env.Client.AddTask(ctx, v.tgt.Prefix, fields)
			action := "add task"
			if err == nil {
				action = res.Action + " " + res.ID
			}
			return formSubmitMsg{form: v, action: action, warnings: res.Warnings, err: err}
		}
		res, err := v.env.Client.Edit(ctx, v.data.checkout.Dir, v.tgt.ID, fields)
		return formSubmitMsg{form: v, action: "edit " + v.tgt.ID, warnings: res.Warnings, err: err}
	}
}

func (v *taskFormView) render(width, height int) string {
	if v.data == nil {
		text := v.env.Styles.Muted.Render("loading " + v.tgt.ID + " …")
		if v.problem != "" {
			text = v.env.Styles.Error.Render(v.problem)
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, text)
	}
	formWidth := min(68, max(20, width-8))
	v.titleInput.SetWidth(formWidth)
	v.body.SetWidth(formWidth)
	v.tags.SetWidth(formWidth)
	s := v.env.Styles
	content := s.Header.Render(v.titleText()) + s.Muted.Render("   ctrl+s save · esc cancel · tab next") +
		"\n\n" + s.Muted.Render("title") + "\n" + v.titleInput.View() +
		"\n\n" + s.Muted.Render("body") + "\n" + v.body.View() +
		"\n\n" + s.Muted.Render("tags") + "\n" + v.tags.View()
	if v.problem != "" {
		content += "\n\n" + s.Error.Render(v.problem)
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, s.Frame.Render(content))
}

func (a *App) openEdit() tea.Cmd {
	tgt := a.selected()
	if tgt == nil {
		return nil
	}
	return a.push(newTaskFormView(a.env, *tgt))
}

func (a *App) openAdd() tea.Cmd {
	prefix := a.top().project()
	if prefix == "" {
		a.log.Add(LevelInfo, "no project highlighted")
		return nil
	}
	return a.push(newAddFormView(a.env, prefix))
}
