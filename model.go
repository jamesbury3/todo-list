package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func isSpecialKey(key string) bool {
	// Filter out special keys and control sequences that shouldn't be added to text input
	specialKeys := []string{
		"ctrl+c", "ctrl+d", "ctrl+z", "ctrl+a", "ctrl+e", "ctrl+k", "ctrl+u",
		"up", "down", "left", "right",
		"home", "end", "pgup", "pgdown",
		"delete", "insert",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
		"alt+enter", "shift+enter",
	}

	for _, sk := range specialKeys {
		if key == sk {
			return true
		}
	}

	// Filter out other control sequences (keys starting with ctrl+, alt+, etc.)
	if strings.HasPrefix(key, "ctrl+") || strings.HasPrefix(key, "alt+") || strings.HasPrefix(key, "meta+") {
		return true
	}

	return false
}

func initialModel() model {
	m := model{
		inProgress:  loadTodos(inProgressFile),
		completed:   loadTodos(completedFile),
		cursor:      0,
		currentView: viewInProgress,
	}
	m.updateDisplayedCompleted()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.adding {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.newTodo) != "" {
					newTodo := Todo{
						Text:      m.newTodo,
						CreatedAt: time.Now(),
					}
					m.inProgress = append(m.inProgress, newTodo)
					saveTodos(inProgressFile, m.inProgress)
					m.message = "Todo added!"
				}
				m.adding = false
				m.newTodo = ""
			case "esc":
				m.adding = false
				m.newTodo = ""
				m.message = "Cancelled"
			case "backspace":
				if len(m.newTodo) > 0 {
					// Handle UTF-8 properly by converting to runes
					runes := []rune(m.newTodo)
					m.newTodo = string(runes[:len(runes)-1])
				}
			default:
				// Filter out special keys but allow pasting and special characters
				key := msg.String()
				// Ignore control sequences and special keys
				if !isSpecialKey(key) {
					// Strip bracketed paste markers if present
					key = strings.TrimPrefix(key, "[")
					key = strings.TrimSuffix(key, "]")
					m.newTodo += key
				}
			}
			return m, nil
		}

		if m.editingDescription {
			switch msg.String() {
			case "enter":
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					if m.currentView == viewInProgress {
						m.inProgress[m.cursor].Description = m.newDescription
						saveTodos(inProgressFile, m.inProgress)
					} else {
						// Find and update in completed list
						todoToUpdate := m.displayedCompleted[m.cursor]
						for i, todo := range m.completed {
							if todo.Text == todoToUpdate.Text && todo.CreatedAt.Equal(todoToUpdate.CreatedAt) {
								m.completed[i].Description = m.newDescription
								break
							}
						}
						m.updateDisplayedCompleted()
						saveTodos(completedFile, m.completed)
					}
					m.message = "Description updated!"
					m.showingDescription = true
				}
				m.editingDescription = false
				m.newDescription = ""
			case "esc":
				m.editingDescription = false
				m.newDescription = ""
				m.message = "Cancelled"
			case "backspace":
				if len(m.newDescription) > 0 {
					// Handle UTF-8 properly by converting to runes
					runes := []rune(m.newDescription)
					m.newDescription = string(runes[:len(runes)-1])
				}
			default:
				// Filter out special keys but allow pasting and special characters
				key := msg.String()
				// Ignore control sequences and special keys
				if !isSpecialKey(key) {
					// Strip bracketed paste markers if present
					key = strings.TrimPrefix(key, "[")
					key = strings.TrimSuffix(key, "]")
					m.newDescription += key
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j":
			currentList := m.getCurrentList()
			if m.cursor < len(currentList)-1 {
				m.cursor++
			}
			m.message = ""
			m.showingDescription = false

		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.message = ""
			m.showingDescription = false

		case "J":
			if m.currentView == viewInProgress && len(m.inProgress) > 0 && m.cursor < len(m.inProgress)-1 {
				// Swap current item with the one below it
				m.inProgress[m.cursor], m.inProgress[m.cursor+1] = m.inProgress[m.cursor+1], m.inProgress[m.cursor]
				m.cursor++
				saveTodos(inProgressFile, m.inProgress)
				m.message = "Todo moved down"
			}

		case "K":
			if m.currentView == viewInProgress && len(m.inProgress) > 0 && m.cursor > 0 {
				// Swap current item with the one above it
				m.inProgress[m.cursor], m.inProgress[m.cursor-1] = m.inProgress[m.cursor-1], m.inProgress[m.cursor]
				m.cursor--
				saveTodos(inProgressFile, m.inProgress)
				m.message = "Todo moved up"
			}

		case "h":
			if m.currentView == viewCompleted {
				m.currentView = viewInProgress
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			}

		case "l":
			if m.currentView == viewInProgress {
				m.currentView = viewCompleted
				m.updateDisplayedCompleted()
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			}

		case "a":
			if m.currentView == viewInProgress {
				m.adding = true
				m.newTodo = ""
				m.message = ""
			}

		case "d":
			if m.currentView == viewInProgress {
				if len(m.inProgress) > 0 && m.cursor < len(m.inProgress) {
					m.inProgress = append(m.inProgress[:m.cursor], m.inProgress[m.cursor+1:]...)
					if m.cursor >= len(m.inProgress) && m.cursor > 0 {
						m.cursor--
					}
					saveTodos(inProgressFile, m.inProgress)
					m.message = "Todo deleted"
				}
			} else {
				if len(m.displayedCompleted) > 0 && m.cursor < len(m.displayedCompleted) {
					// Find and remove from the actual completed list
					todoToDelete := m.displayedCompleted[m.cursor]
					for i, todo := range m.completed {
						if todo.Text == todoToDelete.Text && todo.CreatedAt.Equal(todoToDelete.CreatedAt) {
							m.completed = append(m.completed[:i], m.completed[i+1:]...)
							break
						}
					}
					m.updateDisplayedCompleted()
					if m.cursor >= len(m.displayedCompleted) && m.cursor > 0 {
						m.cursor--
					}
					saveTodos(completedFile, m.completed)
					m.message = "Todo deleted"
				}
			}

		case "x":
			if m.currentView == viewInProgress && len(m.inProgress) > 0 && m.cursor < len(m.inProgress) {
				todo := m.inProgress[m.cursor]
				now := time.Now()
				todo.CompletedAt = &now
				m.inProgress = append(m.inProgress[:m.cursor], m.inProgress[m.cursor+1:]...)
				m.completed = append(m.completed, todo)
				m.updateDisplayedCompleted()
				if m.cursor >= len(m.inProgress) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(inProgressFile, m.inProgress)
				saveTodos(completedFile, m.completed)
				m.message = "Todo completed!"
			}

		case "u":
			if m.currentView == viewCompleted && len(m.displayedCompleted) > 0 && m.cursor < len(m.displayedCompleted) {
				// Find and remove from the actual completed list
				todoToUndo := m.displayedCompleted[m.cursor]
				for i, todo := range m.completed {
					if todo.Text == todoToUndo.Text && todo.CreatedAt.Equal(todoToUndo.CreatedAt) {
						// Clear the completion timestamp
						todoToUndo.CompletedAt = nil
						m.completed = append(m.completed[:i], m.completed[i+1:]...)
						m.inProgress = append(m.inProgress, todoToUndo)
						break
					}
				}
				m.updateDisplayedCompleted()
				if m.cursor >= len(m.displayedCompleted) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(inProgressFile, m.inProgress)
				saveTodos(completedFile, m.completed)
				m.message = "Todo moved to in-progress!"
			}

		case "i":
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				m.showingDescription = !m.showingDescription
				m.message = ""
			}

		case "I":
			m.showingAllDescriptions = !m.showingAllDescriptions
			m.message = ""

		case "e":
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				m.editingDescription = true
				m.newDescription = currentList[m.cursor].Description
				m.message = ""
			}
		}
	}

	return m, nil
}

func (m *model) getCurrentList() []Todo {
	if m.currentView == viewInProgress {
		return m.inProgress
	}
	return m.displayedCompleted
}

func (m *model) updateDisplayedCompleted() {
	if len(m.completed) == 0 {
		m.displayedCompleted = []Todo{}
		return
	}

	// Sort by completion time (most recent first)
	sorted := make([]Todo, len(m.completed))
	copy(sorted, m.completed)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CompletedAt == nil {
			return false
		}
		if sorted[j].CompletedAt == nil {
			return true
		}
		return sorted[i].CompletedAt.After(*sorted[j].CompletedAt)
	})

	// Take only the first 10
	if len(sorted) > 10 {
		m.displayedCompleted = sorted[:10]
	} else {
		m.displayedCompleted = sorted
	}
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("\n")

	if m.currentView == viewInProgress {
		s.WriteString("  [IN PROGRESS]  Completed (l to switch)\n\n")
	} else {
		s.WriteString("  In Progress (h to switch)  [COMPLETED (10 most recent)]\n\n")
	}

	currentList := m.getCurrentList()

	if len(currentList) == 0 {
		s.WriteString("  No todos\n")
	} else {
		for i, todo := range currentList {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			// Add description indicator
			indicator := ""
			if todo.Description != "" {
				indicator = " 📄"
			}

			// Format the display based on view
			if m.currentView == viewCompleted && todo.CompletedAt != nil {
				completedTime := todo.CompletedAt.Format("Jan 2, 15:04")
				s.WriteString(fmt.Sprintf("  %s %s%s [%s]\n", cursor, todo.Text, indicator, completedTime))
			} else {
				createdTime := todo.CreatedAt.Format("Jan 2, 15:04")
				s.WriteString(fmt.Sprintf("  %s %s%s [%s]\n", cursor, todo.Text, indicator, createdTime))
			}

			// Show description if toggled and cursor is on this todo, or if showing all descriptions
			if todo.Description != "" && ((m.showingDescription && i == m.cursor) || m.showingAllDescriptions) {
				s.WriteString(fmt.Sprintf("     └─ %s\n", todo.Description))
			}
		}
	}

	s.WriteString("\n")

	if m.adding {
		s.WriteString("  Add new todo: " + m.newTodo + "_\n")
		s.WriteString("  (press Enter to save, Esc to cancel)\n\n")
	} else if m.editingDescription {
		s.WriteString("  Edit description: " + m.newDescription + "_\n")
		s.WriteString("  (press Enter to save, Esc to cancel)\n\n")
	} else {
		s.WriteString("  Commands:\n")
		s.WriteString("  j/k: move down/up  J/K: reorder (in progress only)  h/l: switch views\n")
		s.WriteString("  a: add  d: delete  x: mark complete  u: undo complete\n")
		s.WriteString("  i: toggle description  I: toggle all descriptions  e: edit description  q: quit\n\n")
	}

	if m.message != "" {
		s.WriteString("  " + m.message + "\n")
	}

	return s.String()
}
