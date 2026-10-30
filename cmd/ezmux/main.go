package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/nwassom/ezmux/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.New()).Run(); err != nil {
		log.Fatal(err)
	}
}
