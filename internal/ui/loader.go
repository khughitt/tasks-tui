package ui

import (
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
)

// loadMsg carries a fetch result back to the view whose loader issued it.
type loadMsg struct {
	loader     uint64
	gen        uint64
	data       any
	err        error
	background bool
}

var loaderIDs atomic.Uint64

// Loader is spec §10: reloads coalesce into at most one pending request, and a result
// older than the latest generation is discarded.
type Loader struct {
	id       uint64
	gen      uint64
	inflight bool
	pending  bool
	next     func(gen uint64) tea.Cmd
}

func NewLoader() *Loader { return &Loader{id: loaderIDs.Add(1)} }

func (l *Loader) ID() uint64     { return l.id }
func (l *Loader) InFlight() bool { return l.inflight }

// Request runs run(gen) now, or records it to run once the in-flight load returns.
func (l *Loader) Request(run func(gen uint64) tea.Cmd) tea.Cmd {
	l.gen++
	l.next = run
	if l.inflight {
		l.pending = true
		return nil
	}
	l.inflight = true
	return run(l.gen)
}

// Done reports whether a result for gen is the latest, and hands back the pending
// reload when there is one.
func (l *Loader) Done(gen uint64) (accept bool, next tea.Cmd) {
	l.inflight = false
	if l.pending {
		l.pending = false
		l.inflight = true
		next = l.next(l.gen)
	}
	return gen == l.gen, next
}

// Cmd wraps a fetch so its result routes back to this loader at this generation.
func (l *Loader) Cmd(gen uint64, fetch func() (any, error)) tea.Cmd {
	id := l.id
	return func() tea.Msg {
		data, err := fetch()
		return loadMsg{loader: id, gen: gen, data: data, err: err}
	}
}
