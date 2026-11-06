package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/nwassom/ezmux/internal/tmux"
)

var (
	accent      = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	dim         = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	muted       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selected    = lipgloss.NewStyle().Background(lipgloss.Color("24")).Foreground(lipgloss.Color("255")).Bold(true)
	panelBorder = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
)

func (m Model) View() tea.View {
	width := max(32, m.width-2)
	height := max(8, m.height)
	background := lipgloss.Color("#1e1e2e")
	if m.err != nil && len(m.sessions) == 0 {
		content := strings.Join([]string{"ezmux", "", m.err.Error(), "", muted.Render("esc/q to close")}, "\n")
		view := tea.NewView(fillBackground(content, width, background))
		view.BackgroundColor = background
		return view
	}

	start, end, windowOffset, windowRows := m.windowViewport()
	previewRows := max(1, height-windowRows-6)
	windows := m.windows()

	heading := fmt.Sprintf("WINDOWS — %s (%d)", m.selectedSessionName(), len(windows))
	if start > 0 {
		heading += " ↑"
	}
	if end < len(windows) {
		heading += " ↓"
	}
	lines := []string{
		center(ansi.Truncate(m.sessionTabs(width), width, "…"), width),
		center(accent.Render(clip(heading, width)), width),
	}
	windowWidth := min(60, width)
	for row := 0; row < windowRows; row++ {
		windowLine := ""
		index := start + row - windowOffset
		if index >= start && index < end {
			window := windows[index]
			windowLine = fmt.Sprintf("%s %-3d  %s", marker(index == m.windowCursor), window.Index, window.Name)
			if window.Active {
				windowLine += dim.Render("  active")
			}
			windowLine = ansi.Truncate(windowLine, windowWidth, "…")
			windowLine += strings.Repeat(" ", max(0, windowWidth-lipgloss.Width(windowLine)))
			if index == m.windowCursor {
				windowLine = selected.Render(windowLine)
			}
		} else if len(windows) == 0 && row == windowRows/2 {
			windowLine = muted.Render("No windows in this session.")
		}
		lines = append(lines, center(windowLine, width))
	}
	lines = append(lines, center(accent.Render(m.previewTitle()), width))
	panelWidth := max(20, width-8)
	innerWidth := panelWidth - 4
	lines = append(lines, center(panelBorder.Render("╭"+strings.Repeat("─", panelWidth-2)+"╮"), width))
	for _, line := range renderPaneLayout(m.preview, m.previewErr, previewRows, innerWidth) {
		lines = append(lines, center(frameLine(line, panelWidth), width))
	}
	lines = append(lines, center(panelBorder.Render("╰"+strings.Repeat("─", panelWidth-2)+"╯"), width))
	lines = append(lines, center(ansi.Truncate(muted.Render("h/l sessions · j/k windows · enter switch · mouse hover/click · esc quit"), width, "…"), width))

	view := tea.NewView(fillBackground(strings.Join(lines, "\n"), width, background))
	view.BackgroundColor = background
	view.MouseMode = tea.MouseModeAllMotion
	return view
}

func (m Model) sessionTabs(width int) string {
	if len(m.sessions) == 0 {
		return muted.Render("no tmux sessions")
	}
	labels := make([]string, 0, len(m.sessions))
	for i, label := range m.sessionLabels(width) {
		if i == m.sessionCursor {
			labels = append(labels, selected.Render(label))
		} else {
			labels = append(labels, muted.Render(label))
		}
	}
	return strings.Join(labels, "  ")
}

func (m Model) sessionLabels(width int) []string {
	labels := make([]string, len(m.sessions))
	for i, session := range m.sessions {
		labels[i] = fmt.Sprintf(" %s ", clip(session.Name, max(8, width/len(m.sessions)-2)))
	}
	return labels
}

func (m Model) sessionAt(x int) int {
	width := max(32, m.width-2)
	labels := m.sessionLabels(width)
	totalWidth := max(0, len(labels)-1) * 2
	for _, label := range labels {
		totalWidth += lipgloss.Width(label)
	}
	left := max(0, (width-totalWidth)/2)
	if totalWidth > width {
		left = 0
	}
	for i, label := range labels {
		right := left + lipgloss.Width(label)
		if x >= left && x < right && x < width {
			return i
		}
		left = right + 2
	}
	return -1
}

func (m Model) windowAt(y int) int {
	start, end, offset, rows := m.windowViewport()
	row := y - 2
	if row < 0 || row >= rows {
		return -1
	}
	index := start + row - offset
	if index < start || index >= end {
		return -1
	}
	return index
}

func (m Model) windowViewport() (start, end, offset, rows int) {
	rows = min(10, max(1, (max(8, m.height)-8)/3))
	windows := m.windows()
	start, end = visibleWindowRange(len(windows), m.windowCursor, rows)
	if m.windowScrollFrozen {
		start = min(max(0, m.windowScrollStart), max(0, len(windows)-rows))
		end = min(len(windows), start+rows)
	}
	offset = max(0, (rows-(end-start))/2)
	return
}

func (m Model) previewTitle() string {
	_, window, ok := m.selection()
	if !ok {
		return "PREVIEW"
	}
	return "PREVIEW — " + window.Name
}

func (m Model) selectedSessionName() string {
	if m.sessionCursor < 0 || m.sessionCursor >= len(m.sessions) {
		return ""
	}
	return m.sessions[m.sessionCursor].Name
}

func renderPaneLayout(panes []tmux.PanePreview, captureErr error, height, width int) []string {
	if captureErr != nil {
		return fillLines(muted.Render("Preview unavailable: "+captureErr.Error()), height)
	}
	if len(panes) == 0 {
		return fillLines(muted.Render("Select a window to preview its panes."), height)
	}

	totalWidth, totalHeight := 0, 0
	for _, pane := range panes {
		totalWidth = max(totalWidth, pane.Pane.Left+pane.Pane.Width)
		totalHeight = max(totalHeight, pane.Pane.Top+pane.Pane.Height)
	}
	if totalWidth == 0 || totalHeight == 0 {
		return fillLines(clipAnsi(panes[0].Content, width), height)
	}

	visuals := make([]paneVisual, 0, len(panes))
	for _, preview := range panes {
		bounds := scalePane(preview.Pane, totalWidth, totalHeight, width, height)
		content := strings.Split(strings.TrimRight(preview.Content, "\n"), "\n")
		if len(content) == 1 && content[0] == "" {
			content = nil
		}
		if preview.Pane.Height > 0 && len(content) > preview.Pane.Height {
			content = content[len(content)-preview.Pane.Height:]
		}
		visuals = append(visuals, paneVisual{preview: preview, bounds: bounds, content: content})
	}
	sort.Slice(visuals, func(i, j int) bool { return visuals[i].bounds.left < visuals[j].bounds.left })

	rows := make([]string, height)
	for y := 0; y < height; y++ {
		var row strings.Builder
		cursor, priorRight := 0, -1
		for _, visual := range visuals {
			if y < visual.bounds.top || y >= visual.bounds.bottom {
				continue
			}
			left := visual.bounds.left
			if priorRight >= 0 && visual.preview.Pane.Left >= priorRight && cursor < width {
				row.WriteString(panelBorder.Render("│"))
				cursor++
			}
			if left < cursor {
				left = cursor
			}
			if left > cursor {
				row.WriteString(strings.Repeat(" ", left-cursor))
				cursor = left
			}
			right := min(width, visual.bounds.right)
			cellWidth := max(0, right-cursor)
			if cellWidth == 0 {
				continue
			}
			localY := y - visual.bounds.top
			var text string
			if localY == 0 {
				label := fmt.Sprintf("Pane %d · %s", visual.preview.Pane.Index, visual.preview.Pane.Command)
				if visual.preview.Pane.Active {
					text = accent.Render(clip(label, cellWidth))
				} else {
					text = dim.Render(clip(label, cellWidth))
				}
			} else if line := localY - 1; line < len(visual.content) {
				text = clipAnsi(visual.content[line], cellWidth)
			}
			row.WriteString(text)
			cursor += lipgloss.Width(text)
			if padding := cellWidth - lipgloss.Width(text); padding > 0 {
				row.WriteString(strings.Repeat(" ", padding))
				cursor += padding
			}
			priorRight = visual.preview.Pane.Left + visual.preview.Pane.Width
		}
		if cursor < width {
			row.WriteString(strings.Repeat(" ", width-cursor))
		}
		rows[y] = row.String()
	}
	return rows
}

type paneVisual struct {
	preview tmux.PanePreview
	bounds  paneBounds
	content []string
}

type paneBounds struct{ left, top, right, bottom int }

func scalePane(pane tmux.Pane, totalWidth, totalHeight, width, height int) paneBounds {
	return paneBounds{
		left:   pane.Left * width / totalWidth,
		top:    pane.Top * height / totalHeight,
		right:  (pane.Left + pane.Width) * width / totalWidth,
		bottom: (pane.Top + pane.Height) * height / totalHeight,
	}
}

func fillLines(first string, height int) []string {
	lines := make([]string, height)
	if height > 0 {
		lines[0] = first
	}
	return lines
}

func frameLine(content string, width int) string {
	content = ansi.Truncate(content, width-4, "…")
	line := panelBorder.Render("│") + "  " + content
	line += strings.Repeat(" ", max(0, width-1-lipgloss.Width(line))) + panelBorder.Render("│")
	return line
}

func visibleWindowRange(total, cursor, visible int) (int, int) {
	if total <= visible {
		return 0, total
	}
	start := max(0, min(cursor-visible/2, total-visible))
	return start, start + visible
}

func center(value string, width int) string {
	padding := max(0, (width-lipgloss.Width(value))/2)
	return strings.Repeat(" ", padding) + value
}

func fillBackground(content string, width int, background color.Color) string {
	style := lipgloss.NewStyle().Background(background).Width(width)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = style.Render(ansi.Truncate(line, width, "…"))
	}
	return strings.Join(lines, "\n")
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
