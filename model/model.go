package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

// InitialModel creates and returns the initial model state
func InitialModel() Model {
	m := Model{
		backlog:     loadTodos(backlogFile),
		ready:       loadTodos(readyFile),
		completed:   loadTodos(completedFile),
		cursor:      0,
		currentView: viewReady,
	}
	m.updateDisplayedCompleted()
	return m
}

// Init initializes the model and returns the initial command
func (m Model) Init() tea.Cmd {
	return tea.ClearScreen
}
