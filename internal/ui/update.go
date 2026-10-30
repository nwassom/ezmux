package ui

import (
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nwassom/ezmux/internal/tmux"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case sessionsMsg:
		m.err, m.sessions = msg.err, msg.sessions
		m.sessionCursor = min(m.sessionCursor, max(0, len(m.sessions)-1))
		m.windowCursor = min(m.windowCursor, max(0, len(m.windows())-1))
		return m, m.refreshPreview()
	case previewMsg:
		m.preview = msg.content
	case tickMsg:
		return m, tea.Batch(load(m.client), tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))
	case switchMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, tea.Quit
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			m.sessionCursor = wrap(m.sessionCursor-1, len(m.sessions))
			m.windowCursor = 0
			return m, m.refreshPreview()
		case "right", "l":
			m.sessionCursor = wrap(m.sessionCursor+1, len(m.sessions))
			m.windowCursor = 0
			return m, m.refreshPreview()
		case "up", "k":
			m.windowCursor = wrap(m.windowCursor-1, len(m.windows()))
			return m, m.refreshPreview()
		case "down", "j":
			m.windowCursor = wrap(m.windowCursor+1, len(m.windows()))
			return m, m.refreshPreview()
		case "enter":
			if session, window, ok := m.selection(); ok {
				return m, func() tea.Msg { return switchMsg{err: m.client.Switch(session, window)} }
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			index, _ := strconv.Atoi(msg.String())
			for i, window := range m.windows() {
				if window.Index == index {
					m.windowCursor = i
					return m, m.refreshPreview()
				}
			}
		}
	}
	return m, nil
}

type switchMsg struct{ err error }

func load(client tmux.Client) tea.Cmd {
	return func() tea.Msg { sessions, err := client.Sessions(); return sessionsMsg{sessions, err} }
}
func (m Model) refreshPreview() tea.Cmd {
	_, window, ok := m.selection()
	if !ok {
		return nil
	}
	return func() tea.Msg { content, _ := m.client.Capture(window); return previewMsg{content} }
}
func (m Model) windows() []tmux.Window {
	if m.sessionCursor < 0 || m.sessionCursor >= len(m.sessions) {
		return nil
	}
	return m.sessions[m.sessionCursor].Windows
}
func (m Model) selection() (tmux.Session, tmux.Window, bool) {
	windows := m.windows()
	if len(windows) == 0 {
		return tmux.Session{}, tmux.Window{}, false
	}
	return m.sessions[m.sessionCursor], windows[m.windowCursor], true
}
func wrap(value, length int) int {
	if length == 0 {
		return 0
	}
	if value < 0 {
		return length - 1
	}
	if value >= length {
		return 0
	}
	return value
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
