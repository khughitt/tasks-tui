package ui

// The message log of spec §11: every error and warning of the session, and the one
// entry the status line shows until the next keypress.

type Level int

const (
	LevelInfo Level = iota
	LevelWarning
	LevelError
)

type Message struct {
	Level Level
	Text  string
}

type Log struct {
	Entries []Message
	current *Message
}

// Add records a message and shows it unless a more severe one is already showing.
func (l *Log) Add(level Level, text string) {
	l.Entries = append(l.Entries, Message{level, text})
	if l.current == nil || level >= l.current.Level {
		m := l.Entries[len(l.Entries)-1]
		l.current = &m
	}
}

// AddWarnings records a successful command's warnings[], each prefixed with the command.
func (l *Log) AddWarnings(command string, warnings []string) {
	for _, w := range warnings {
		l.Add(LevelWarning, command+": "+w)
	}
}

func (l *Log) Current() *Message { return l.current }

// Dismiss clears the status line; the log keeps everything.
func (l *Log) Dismiss() { l.current = nil }
