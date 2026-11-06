package tmux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// tmux escapes control characters in format output, so this separator arrives as the text "\037".
const fieldSeparator = "\\037"

type Session struct {
	ID, Name string
	Windows  []Window
}
type Window struct {
	ID, SessionID string
	Index         int
	Name          string
	Active        bool
	ActivePaneID  string
	Panes         []Pane
}
type Pane struct {
	ID, WindowID, Command string
	Index, Width, Height  int
	Left, Top             int
	Active                bool
}
type PanePreview struct {
	Pane    Pane
	Content string
}
type Client struct{}

func (Client) ResizePopup() error {
	if os.Getenv("EZMUX_POPUP") != "1" {
		return nil
	}
	paneID := os.Getenv("TMUX_PANE")
	if paneID == "" {
		return nil
	}
	out, err := run("display-message", "-p", "-t", paneID, "#{pane_width} #{pane_height} #{window_width} #{window_height}")
	if err != nil {
		return err
	}
	fields := strings.Fields(out)
	if len(fields) != 4 {
		return fmt.Errorf("unexpected popup dimensions: %q", out)
	}
	dimensions := [4]int{}
	for i, field := range fields {
		dimensions[i], err = strconv.Atoi(field)
		if err != nil {
			return err
		}
	}
	width, height := popupSize(dimensions[2], dimensions[3])
	if dimensions[0] == width && dimensions[1] == height {
		return nil
	}
	_, err = run("resize-pane", "-t", paneID, "-x", strconv.Itoa(width), "-y", strconv.Itoa(height))
	return err
}

func popupSize(width, height int) (int, int) {
	return max(1, width*85/100), max(1, height*80/100)
}

func (Client) Sessions() ([]Session, error) {
	out, err := run("list-sessions", "-F", "#{session_id}"+fieldSeparator+"#{session_name}")
	if err != nil {
		return nil, err
	}
	result := ParseSessions(out)
	for i := range result {
		result[i].Windows, err = windows(result[i].ID)
		if err != nil {
			return nil, err
		}
	}
	allPanes, err := panes()
	if err != nil {
		return nil, err
	}
	panesByWindow := make(map[string][]Pane)
	for _, pane := range allPanes {
		panesByWindow[pane.WindowID] = append(panesByWindow[pane.WindowID], pane)
	}
	for i := range result {
		for j := range result[i].Windows {
			result[i].Windows[j].Panes = panesByWindow[result[i].Windows[j].ID]
		}
	}
	return result, nil
}

func ParseSessions(output string) []Session {
	var result []Session
	for _, line := range lines(output) {
		fields := strings.Split(line, fieldSeparator)
		if len(fields) != 2 {
			continue
		}
		result = append(result, Session{ID: fields[0], Name: fields[1]})
	}
	return result
}

func (Client) Capture(window Window) ([]PanePreview, error) {
	panes := window.Panes
	if len(panes) == 0 && window.ActivePaneID != "" {
		panes = []Pane{{ID: window.ActivePaneID, Active: true}}
	}
	previews := make([]PanePreview, 0, len(panes))
	for _, pane := range panes {
		content, err := run("capture-pane", "-p", "-e", "-S", "-100", "-t", pane.ID)
		if err != nil {
			return nil, err
		}
		previews = append(previews, PanePreview{Pane: pane, Content: content})
	}
	return previews, nil
}
func (Client) Switch(session Session, window Window) error {
	_, err := run("switch-client", "-t", session.ID+":"+window.ID)
	return err
}

func windows(sessionID string) ([]Window, error) {
	format := strings.Join([]string{"#{window_id}", "#{session_id}", "#{window_index}", "#{window_name}", "#{window_active}", "#{pane_id}"}, fieldSeparator)
	out, err := run("list-windows", "-t", sessionID, "-F", format)
	if err != nil {
		return nil, err
	}
	return ParseWindows(out), nil
}

func panes() ([]Pane, error) {
	fields := []string{"#{window_id}", "#{pane_id}", "#{pane_index}", "#{pane_active}", "#{pane_width}", "#{pane_height}", "#{pane_left}", "#{pane_top}", "#{pane_current_command}"}
	format := strings.Join(fields, fieldSeparator)
	out, err := run("list-panes", "-a", "-F", format)
	if err != nil {
		return nil, err
	}
	return ParsePanes(out), nil
}

func ParsePanes(output string) []Pane {
	var result []Pane
	for _, line := range lines(output) {
		fields := strings.Split(line, fieldSeparator)
		if len(fields) != 9 {
			continue
		}
		values := [5]int{}
		valid := true
		for i, fieldIndex := range []int{2, 4, 5, 6, 7} {
			value, err := strconv.Atoi(fields[fieldIndex])
			if err != nil {
				valid = false
				break
			}
			values[i] = value
		}
		if !valid {
			continue
		}
		result = append(result, Pane{WindowID: fields[0], ID: fields[1], Index: values[0], Active: fields[3] == "1", Width: values[1], Height: values[2], Left: values[3], Top: values[4], Command: fields[8]})
	}
	return result
}

func ParseWindows(output string) []Window {
	var result []Window
	for _, line := range lines(output) {
		fields := strings.Split(line, fieldSeparator)
		if len(fields) != 6 {
			continue
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		result = append(result, Window{ID: fields[0], SessionID: fields[1], Index: index, Name: fields[3], Active: fields[4] == "1", ActivePaneID: fields[5]})
	}
	return result
}

func run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("tmux: %s", message)
		}
		return "", err
	}
	return string(out), nil
}

func lines(output string) []string {
	output = strings.TrimRight(output, "\r\n")
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
}
