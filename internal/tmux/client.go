package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const fieldSeparator = "\x1f"

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
}
type Client struct{}

func (Client) Sessions() ([]Session, error) {
	out, err := run("list-sessions", "-F", "#{session_id}"+fieldSeparator+"#{session_name}")
	if err != nil {
		return nil, err
	}
	var result []Session
	for _, line := range lines(out) {
		fields := strings.Split(line, fieldSeparator)
		if len(fields) != 2 {
			continue
		}
		windows, err := windows(fields[0])
		if err != nil {
			return nil, err
		}
		result = append(result, Session{ID: fields[0], Name: fields[1], Windows: windows})
	}
	return result, nil
}

func (Client) Capture(window Window) (string, error) {
	return run("capture-pane", "-p", "-e", "-S", "-100", "-t", window.ActivePaneID)
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
