package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nwassom/ezmux/internal/tmux"
)

type sessionsMsg struct {
	sessions []tmux.Session
	err      error
}
type previewMsg struct{ content string }
type tickMsg time.Time

type Model struct {
	client                      tmux.Client
	sessions                    []tmux.Session
	sessionCursor, windowCursor int
	preview                     string
	err                         error
	width, height               int
}

func New() Model { return Model{client: tmux.Client{}} }
func (m Model) Init() tea.Cmd {
	return tea.Batch(load(m.client), tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))
}
