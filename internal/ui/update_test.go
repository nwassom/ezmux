package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nwassom/ezmux/internal/tmux"
)

func TestNavigation(t *testing.T) {
	m := Model{sessions: []tmux.Session{{ID: "$1", Windows: []tmux.Window{{Index: 0}, {Index: 2}}}, {ID: "$2", Windows: []tmux.Window{{Index: 1}}}}}
	updated, _ := m.Update(tea.KeyPressMsg{Text: "l", Code: 'l'})
	m = updated.(Model)
	if m.sessionCursor != 1 || m.windowCursor != 0 {
		t.Fatalf("right navigation: session=%d window=%d", m.sessionCursor, m.windowCursor)
	}
	updated, _ = m.Update(tea.KeyPressMsg{Text: "0", Code: '0'})
	m = updated.(Model)
	if m.windowCursor != 0 {
		t.Fatalf("number selection moved to %d", m.windowCursor)
	}
	updated, _ = m.Update(tea.KeyPressMsg{Text: "h", Code: 'h'})
	m = updated.(Model)
	if m.sessionCursor != 0 || m.windowCursor != 0 {
		t.Fatalf("left navigation: session=%d window=%d", m.sessionCursor, m.windowCursor)
	}
}

func TestWindowResizeReflowsView(t *testing.T) {
	m := Model{sessions: []tmux.Session{{Name: "work", Windows: []tmux.Window{{Index: 1, Name: "shell"}}}}}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(Model)
	if m.width != 100 || m.height != 40 || strings.Count(m.View().Content, "\n") != 39 {
		t.Fatalf("large layout did not reflow: %dx%d", m.width, m.height)
	}

	updated, _ = m.Update(tea.WindowSizeMsg{Width: 50, Height: 16})
	m = updated.(Model)
	if m.width != 50 || m.height != 16 || strings.Count(m.View().Content, "\n") != 15 {
		t.Fatalf("small layout did not reflow: %dx%d", m.width, m.height)
	}
}

func TestMouseHoverSelectsSessionAndWindowPreview(t *testing.T) {
	m := Model{
		width: 100, height: 20,
		sessions: []tmux.Session{
			{Name: "first", Windows: []tmux.Window{{Index: 0, Name: "shell"}}},
			{Name: "second", Windows: []tmux.Window{{Index: 1, Name: "logs"}, {Index: 2, Name: "editor"}}},
		},
	}
	updated, cmd := m.Update(tea.MouseMotionMsg{X: 50, Y: 0})
	m = updated.(Model)
	if m.sessionCursor != 1 || m.windowCursor != 0 || cmd == nil {
		t.Fatalf("hover did not select session and refresh its preview: session=%d window=%d cmd=%v", m.sessionCursor, m.windowCursor, cmd)
	}

	updated, cmd = m.Update(tea.MouseMotionMsg{X: 50, Y: 4})
	m = updated.(Model)
	if m.windowCursor != 1 || !m.windowScrollFrozen || cmd == nil {
		t.Fatalf("hover did not select window and preview it: window=%d frozen=%v cmd=%v", m.windowCursor, m.windowScrollFrozen, cmd)
	}
	updated, cmd = m.Update(tea.MouseMotionMsg{X: 51, Y: 4})
	m = updated.(Model)
	if m.windowCursor != 1 || cmd != nil {
		t.Fatalf("hover moved the row under the pointer or refreshed redundantly: window=%d cmd=%v", m.windowCursor, cmd)
	}
}

func TestMouseClickSelectsSessionAndSwitchesWindow(t *testing.T) {
	m := Model{
		width: 100, height: 20,
		sessions: []tmux.Session{
			{Name: "first", Windows: []tmux.Window{{ID: "@1", Index: 0, Name: "shell"}}},
			{Name: "second", ID: "$2", Windows: []tmux.Window{{ID: "@2", Index: 1, Name: "logs"}, {ID: "@3", Index: 2, Name: "editor"}}},
		},
	}
	updated, cmd := m.Update(tea.MouseClickMsg{X: 50, Y: 0, Button: tea.MouseLeft})
	m = updated.(Model)
	if m.sessionCursor != 1 || cmd == nil {
		t.Fatalf("click did not select session: cursor=%d cmd=%v", m.sessionCursor, cmd)
	}
	updated, cmd = m.Update(tea.MouseClickMsg{X: 50, Y: 4, Button: tea.MouseLeft})
	m = updated.(Model)
	if m.windowCursor != 1 || cmd == nil {
		t.Fatalf("click did not select and switch to window: cursor=%d cmd=%v", m.windowCursor, cmd)
	}
}
