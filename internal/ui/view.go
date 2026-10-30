package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selected = lipgloss.NewStyle().Background(lipgloss.Color("24")).Foreground(lipgloss.Color("255")).Bold(true)
	border   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

func (m Model) View() tea.View {
	if m.err != nil && len(m.sessions) == 0 {
		return tea.NewView("ezmux\n\n" + m.err.Error() + "\n\n" + muted.Render("esc/q to close"))
	}

	width := max(32, m.width-4)
	windowRows := min(8, max(3, (m.height-12)/3))
	previewRows := max(4, m.height-windowRows-11)

	lines := []string{
		center(accent.Render("EZMUX"), width),
		center(muted.Render("tmux navigator"), width),
		"",
		center(m.sessionTabs(width), width),
		borderLine(width),
	}
	lines = append(lines, m.windowPanel(width, windowRows)...)
	lines = append(lines, "", borderLine(width))
	lines = append(lines, m.previewPanel(width, previewRows)...)
	lines = append(lines, "", center(muted.Render("h/l sessions   j/k windows   0-9 jump   enter switch   esc quit"), width))
	return tea.NewView(strings.Join(lines, "\n"))
}

func (m Model) sessionTabs(width int) string {
	if len(m.sessions) == 0 {
		return muted.Render("no tmux sessions")
	}
	labels := make([]string, 0, len(m.sessions))
	for i, session := range m.sessions {
		label := fmt.Sprintf(" %d  %s ", i+1, clip(session.Name, max(8, width/len(m.sessions)-5)))
		if i == m.sessionCursor {
			labels = append(labels, selected.Render(label))
		} else {
			labels = append(labels, muted.Render(label))
		}
	}
	return strings.Join(labels, "  ")
}

func (m Model) windowPanel(width, height int) []string {
	windows := m.windows()
	header := accent.Render("WINDOWS")
	if len(windows) == 0 {
		return []string{header, muted.Render("  No windows in this session.")}
	}

	start := 0
	if m.windowCursor >= height {
		start = m.windowCursor - height + 1
	}
	end := min(len(windows), start+height)
	more := ""
	if start > 0 {
		more += "  ↑ more"
	}
	if end < len(windows) {
		more += "  ↓ more"
	}
	if more == "" {
		more = fmt.Sprintf("  %d windows", len(windows))
	}
	rows := []string{header + muted.Render(more)}
	for i := start; i < end; i++ {
		window := windows[i]
		name := clip(window.Name, max(8, width-22))
		row := fmt.Sprintf("  %s %-3d  %-*s", marker(i == m.windowCursor), window.Index, max(8, width-22), name)
		if window.Active {
			row += dim.Render(" active")
		}
		if i == m.windowCursor {
			rows = append(rows, selected.Render(row))
		} else {
			rows = append(rows, row)
		}
	}
	for len(rows) < height+1 {
		rows = append(rows, "")
	}
	return rows
}

func (m Model) previewPanel(width, height int) []string {
	content := strings.TrimRight(m.preview, "\n")
	if content == "" {
		content = muted.Render("  Select a window to preview its active pane.")
	}
	preview := strings.Split(content, "\n")
	bodyRows := max(1, height-1)
	if len(preview) > bodyRows {
		preview = preview[len(preview)-bodyRows:]
	}
	for len(preview) < bodyRows {
		preview = append(preview, "")
	}
	result := []string{accent.Render("PREVIEW")}
	for _, line := range preview {
		result = append(result, clipAnsi(line, width))
	}
	return result
}

func borderLine(width int) string {
	return border.Render(strings.Repeat("─", width))
}

func center(value string, width int) string {
	padding := max(0, (width-lipgloss.Width(value))/2)
	return strings.Repeat(" ", padding) + value
}

func clip(value string, width int) string {
	return ansi.Truncate(value, width, "…")
}

func clipAnsi(value string, width int) string {
	return ansi.Truncate(value, width, "…")
}

func marker(isSelected bool) string {
	if isSelected {
		return "›"
	}
	return " "
}
