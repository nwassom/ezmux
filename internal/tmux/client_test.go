package tmux

import (
	"os/exec"
	"testing"
)

func TestClientReadsSessionsAndWindowsFromTmux(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux is not installed")
	}
	if err := exec.Command("tmux", "list-sessions").Run(); err != nil {
		t.Skip("no tmux server is running")
	}

	sessions, err := (Client{}).Sessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) == 0 {
		t.Fatal("tmux sessions were not parsed")
	}
	for _, session := range sessions {
		if len(session.Windows) == 0 || len(session.Windows[0].Panes) == 0 {
			t.Fatalf("session has no parsed windows or pane: %#v", session)
		}
	}
	previews, err := (Client{}).Capture(sessions[0].Windows[0])
	if err != nil || len(previews) != len(sessions[0].Windows[0].Panes) {
		t.Fatalf("captured %d panes, err=%v", len(previews), err)
	}
}

func TestParseSessions(t *testing.T) {
	sessions := ParseSessions("$1\\037work\n$2\\037personal\n")
	if len(sessions) != 2 || sessions[0].ID != "$1" || sessions[0].Name != "work" || sessions[1].Name != "personal" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
}

func TestParseWindows(t *testing.T) {
	windows := ParseWindows("@1\\037$1\\0370\\037main\\0371\\037%2\n@2\\037$1\\0373\\037logs\\0370\\037%4\ninvalid")
	if len(windows) != 2 || windows[1].Index != 3 || windows[1].Active || windows[0].ActivePaneID != "%2" {
		t.Fatalf("unexpected windows: %#v", windows)
	}
}

func TestParsePanes(t *testing.T) {
	panes := ParsePanes("@1\\037%1\\0370\\0371\\03780\\03723\\0370\\0370\\037bash\n@1\\037%2\\0371\\0370\\03740\\03723\\03780\\0370\\037vim")
	if len(panes) != 2 || panes[0].WindowID != "@1" || panes[0].ID != "%1" || !panes[0].Active || panes[0].Index != 0 || panes[0].Width != 80 || panes[1].Command != "vim" || panes[1].Left != 80 {
		t.Fatalf("unexpected panes: %#v", panes)
	}
}
