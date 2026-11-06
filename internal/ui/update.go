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
		m.preview = msg.panes
		m.previewErr = msg.err
	case tickMsg:
		return m, tea.Batch(load(m.client), tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))
	case switchMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, tea.Quit
	case tea.MouseMotionMsg:
		return m.handleMouse(msg.Mouse(), false)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return m.handleMouse(msg.Mouse(), true)
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			m.sessionCursor = wrap(m.sessionCursor-1, len(m.sessions))
			m.windowCursor = 0
			m.windowScrollFrozen = false
			return m, m.refreshPreview()
		case "right", "l":
			m.sessionCursor = wrap(m.sessionCursor+1, len(m.sessions))
			m.windowCursor = 0
			m.windowScrollFrozen = false
			return m, m.refreshPreview()
		case "up", "k":
			m.windowScrollFrozen = false
			m.windowCursor = wrap(m.windowCursor-1, len(m.windows()))
			return m, m.refreshPreview()
		case "down", "j":
			m.windowScrollFrozen = false
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
					m.windowScrollFrozen = false
					m.windowCursor = i
					return m, m.refreshPreview()
				}
			}
		}
	}
	return m, nil
}

func (m Model) handleMouse(mouse tea.Mouse, click bool) (tea.Model, tea.Cmd) {
	if mouse.Y == 0 {
		if session := m.sessionAt(mouse.X); session >= 0 && session != m.sessionCursor {
			m.sessionCursor = session
			m.windowCursor = 0
			m.windowScrollFrozen = false
			return m, m.refreshPreview()
		}
	}
	if index := m.windowAt(mouse.Y); index >= 0 {
		if !click && index == m.windowCursor {
			return m, nil
		}
		if !m.windowScrollFrozen {
			start, _, _, _ := m.windowViewport()
			m.windowScrollStart = start
			m.windowScrollFrozen = true
		}
		m.windowCursor = index
		if click {
			if session, window, ok := m.selection(); ok {
				return m, func() tea.Msg { return switchMsg{err: m.client.Switch(session, window)} }
			}
		} else {
			return m, m.refreshPreview()
		}
	}
	return m, nil
}

type switchMsg struct{ err error }

func load(client tmux.Client) tea.Cmd {
	return func() tea.Msg {
		_ = client.ResizePopup()
		sessions, err := client.Sessions()
		return sessionsMsg{sessions, err}
	}
}
func (m Model) refreshPreview() tea.Cmd {
	_, window, ok := m.selection()
	if !ok {
		return nil
	}
	return func() tea.Msg { panes, err := m.client.Capture(window); return previewMsg{panes, err} }
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
