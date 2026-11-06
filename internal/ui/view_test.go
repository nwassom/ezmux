package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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

func TestViewCentersSelectedWindowAndShowsItsPreview(t *testing.T) {
	m := Model{
		width: 80, height: 20,
		preview: []tmux.PanePreview{
			{Pane: tmux.Pane{Index: 0, Command: "bash", Active: true, Width: 60, Height: 10}, Content: "current terminal output"},
			{Pane: tmux.Pane{Index: 1, Command: "vim", Left: 60, Width: 40, Height: 10}, Content: "second pane output"},
		},
		sessions: []tmux.Session{{Name: "work", Windows: []tmux.Window{{Index: 0, Name: "shell"}, {Index: 1, Name: "logs"}}}},
	}
	view := m.View()
	lines := strings.Split(view.Content, "\n")
	if view.BackgroundColor == nil {
		t.Fatal("view has no opaque background color")
	}
	if view.MouseMode != tea.MouseModeAllMotion {
		t.Fatal("view does not enable mouse hover and clicks")
	}
	if !strings.Contains(lines[0], "work") || !strings.Contains(lines[1], "WINDOWS") || !strings.Contains(lines[6], "PREVIEW") {
		t.Fatalf("session/window headings are missing: %q", lines)
	}
	if strings.Contains(lines[0], "1  work") {
		t.Fatalf("session tab includes a confusing ordinal: %q", lines[0])
	}
	if !strings.Contains(lines[3], "shell") || !strings.Contains(lines[6], "PREVIEW — shell") || !strings.Contains(view.Content, "Pane 0 · bash") || !strings.Contains(view.Content, "Pane 1 · vim") || !strings.Contains(view.Content, "current terminal output") || !strings.Contains(view.Content, "second pane output") {
		t.Fatalf("selected window or terminal preview is missing: %q", view.Content)
	}
	if !strings.Contains(view.Content, "╭") || !strings.Contains(view.Content, "╰") || !strings.Contains(view.Content, "│") {
		t.Fatal("preview panel border is missing")
	}
}

func TestScalePaneKeepsSideBySidePositions(t *testing.T) {
	left := scalePane(tmux.Pane{Left: 0, Top: 0, Width: 60, Height: 30}, 100, 30, 80, 10)
	right := scalePane(tmux.Pane{Left: 60, Top: 0, Width: 40, Height: 30}, 100, 30, 80, 10)
	if left.left != 0 || left.right != 48 || right.left != 48 || right.right != 80 || left.top != right.top || left.bottom != right.bottom {
		t.Fatalf("side-by-side pane bounds overlap or shift: left=%+v right=%+v", left, right)
	}
	top := scalePane(tmux.Pane{Left: 0, Top: 0, Width: 100, Height: 30}, 100, 51, 80, 10)
	bottom := scalePane(tmux.Pane{Left: 0, Top: 31, Width: 100, Height: 20}, 100, 51, 80, 10)
	if top.top != 0 || top.bottom != 5 || bottom.top != 6 || bottom.bottom != 10 {
		t.Fatalf("stacked pane bounds overlap or shift: top=%+v bottom=%+v", top, bottom)
	}
}

func TestWindowNamesAlignAcrossSelectedRows(t *testing.T) {
	m := Model{
		width: 100, height: 20,
		sessions: []tmux.Session{{Name: "work", Windows: []tmux.Window{{Index: 1, Name: "opencodeclear"}, {Index: 2, Name: "Hello World"}}}},
	}
	lines := strings.Split(m.View().Content, "\n")
	selectedLine, otherLine := ansi.Strip(lines[3]), ansi.Strip(lines[4])
	selectedAt, otherAt := strings.Index(selectedLine, "opencodeclear"), strings.Index(otherLine, "Hello World")
	if selectedAt < 0 || otherAt < 0 {
		t.Fatalf("window labels missing: %q / %q", selectedLine, otherLine)
	}
	selectedName := lipgloss.Width(selectedLine[:selectedAt])
	otherName := lipgloss.Width(otherLine[:otherAt])
	if selectedName < 0 || selectedName != otherName {
		t.Fatalf("window names have different horizontal offsets: selected=%d other=%d\nselected: %q\nother: %q", selectedName, otherName, selectedLine, otherLine)
	}
}

func TestVisibleWindowRangeCentersAndScrolls(t *testing.T) {
	start, end := visibleWindowRange(10, 5, 4)
	if start != 3 || end != 7 {
		t.Fatalf("visible range = [%d,%d), want [3,7)", start, end)
	}
	start, end = visibleWindowRange(10, 9, 4)
	if start != 6 || end != 10 {
		t.Fatalf("visible range at end = [%d,%d), want [6,10)", start, end)
	}
}
