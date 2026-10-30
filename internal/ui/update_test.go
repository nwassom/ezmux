package ui

import (
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
