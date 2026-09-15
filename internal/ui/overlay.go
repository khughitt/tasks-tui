package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// promptOverlay is a one-line input with an optional hint line and extra keys.
type promptOverlay struct {
	styles   *Styles
	title    string
	input    textinput.Model
	required string
	hint     func(value string) string
	keys     func(msg tea.KeyPressMsg) bool
	submit   func(value string) (next overlay, cmd tea.Cmd)
	problem  string
}

func newPrompt(s *Styles, title, placeholder string) *promptOverlay {
	in := textinput.New()
	in.Placeholder = placeholder
	in.Prompt = "> "
	in.SetStyles(textinput.DefaultStyles(s.Tone.Dark))
	in.Focus()
	return &promptOverlay{styles: s, title: title, input: in}
}

// update takes every message the App routes to an open overlay: key presses decide
// esc/enter/extra keys, and everything else (a tea.PasteMsg, cursor blinks) goes to
// the text input, which handles pastes itself.
func (o *promptOverlay) update(msg tea.Msg) (overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc":
			return nil, nil
		case "enter":
			value := o.input.Value()
			if o.required != "" && strings.TrimSpace(value) == "" {
				o.problem = o.required
				return o, nil
			}
			return o.submit(value)
		}
		if o.keys != nil && o.keys(k) {
			return o, nil
		}
		o.problem = ""
	}
	var cmd tea.Cmd
	o.input, cmd = o.input.Update(msg)
	return o, cmd
}

func (o *promptOverlay) render(width int) string {
	s := o.styles
	lines := []string{s.Header.Render(o.title) + s.Muted.Render("   enter submit · esc cancel"), o.input.View()}
	if o.problem != "" {
		lines = append(lines, s.Error.Render(o.problem))
	} else if o.hint != nil {
		if h := o.hint(o.input.Value()); h != "" {
			lines = append(lines, h)
		}
	}
	return strings.Join(lines, "\n")
}

// confirmOverlay shows text and runs onAccept on one specific key.
type confirmOverlay struct {
	styles   *Styles
	text     string
	accept   string
	label    string
	onAccept func() tea.Cmd
}

func (o *confirmOverlay) update(msg tea.Msg) (overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return o, nil
	}
	switch k.String() {
	case "esc", "q":
		return nil, nil
	case o.accept:
		return nil, o.onAccept()
	}
	return o, nil
}

func (o *confirmOverlay) render(width int) string {
	s := o.styles
	return o.text + "\n" + s.Bold.Render("["+o.accept+"] "+o.label) + s.Muted.Render("   [esc] cancel")
}

// pickerOverlay chooses one of a few names by j/k+enter or a digit.
type pickerOverlay struct {
	styles *Styles
	title  string
	items  []string
	slot   int
	sel    int
	onPick func(name string) tea.Cmd
}

func (o *pickerOverlay) update(msg tea.Msg) (overlay, tea.Cmd) {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return o, nil
	}
	k := press.String()
	switch {
	case k == "esc" || k == "q":
		return nil, nil
	case k == "enter":
		return nil, o.onPick(o.items[o.sel])
	case k == "j" || k == "down":
		if o.sel < len(o.items)-1 {
			o.sel++
		}
	case k == "k" || k == "up":
		if o.sel > 0 {
			o.sel--
		}
	default:
		if n, err := strconv.Atoi(k); err == nil && n >= 1 && n <= len(o.items) {
			return nil, o.onPick(o.items[n-1])
		}
	}
	return o, nil
}

func (o *pickerOverlay) render(width int) string {
	s := o.styles
	var parts []string
	for i, it := range o.items {
		label := strconv.Itoa(i+1) + " " + it
		if i == o.sel {
			parts = append(parts, s.Gutter(o.slot).Render(" "+label+" "))
		} else {
			parts = append(parts, " "+label+" ")
		}
	}
	return s.Header.Render(o.title) + s.Muted.Render("   enter/digit choose · esc cancel") + "\n" + strings.Join(parts, " ")
}
