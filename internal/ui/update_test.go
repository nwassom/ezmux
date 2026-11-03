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
