package tmux

import "testing"

func TestParseWindows(t *testing.T) {
	windows := ParseWindows("@1\x1f$1\x1f0\x1fmain\x1f1\x1f%2\n@2\x1f$1\x1f3\x1flogs\x1f0\x1f%4\ninvalid")
	if len(windows) != 2 || windows[1].Index != 3 || windows[1].Active || windows[0].ActivePaneID != "%2" {
		t.Fatalf("unexpected windows: %#v", windows)
	}
}
