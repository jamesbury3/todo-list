package main

import (
	"fmt"
	"os"
	"todo-list/model"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(
		model.InitialModel(),
		tea.WithAltScreen(), // Use alternate screen buffer to prevent scrolling
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
