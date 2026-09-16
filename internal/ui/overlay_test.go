package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestTestPromptsDoNotScheduleCursorBlinks(t *testing.T) {
	p := newPrompt(testEnv(newFake()).Styles, "prompt", "")
	_, cmd := p.update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd != nil {
		t.Fatal("test prompt scheduled a cursor blink")
	}
}

func TestPromptsBlinkByDefault(t *testing.T) {
	p := newPrompt(NewStyles(testEnv(newFake()).Styles.Tone), "prompt", "")
	_, cmd := p.update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd == nil {
		t.Fatal("production prompt did not schedule a cursor blink")
	}
}
