package ui

import (
	"strings"
	"testing"

	"github.com/nwassom/ezmux/internal/tmux"
)

func TestViewKeepsGeometryWhileChangingWindows(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		sessions: []tmux.Session{{Name: "work", Windows: []tmux.Window{{Index: 0, Name: "shell"}, {Index: 1, Name: "logs"}}}},
	}
	first := strings.Split(m.View().Content, "\n")
	m.windowCursor = 1
	second := strings.Split(m.View().Content, "\n")
	if len(first) != len(second) {
		t.Fatalf("view height changed from %d to %d", len(first), len(second))
	}
	if !strings.Contains(m.View().Content, "work") {
		t.Fatal("session name is missing from the view")
	}
}
