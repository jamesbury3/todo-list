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

// handleTextInput processes keyboard input for text entry fields
func handleTextInput(key string, currentText *string) bool {
	switch key {
	case "backspace":
		if len(*currentText) > 0 {
			runes := []rune(*currentText)
			*currentText = string(runes[:len(runes)-1])
		}
		return true
	default:
		if !isSpecialKey(key) {
			// Strip bracketed paste markers if present
			key = strings.TrimPrefix(key, "[")
			key = strings.TrimSuffix(key, "]")
			*currentText += key
		}
		return true
	}
}

// updateCompletedTodo finds and updates a todo in the completed list
func (m *model) updateCompletedTodo(updateFn func(*Todo)) {
	if m.cursor >= len(m.displayedCompleted) {
		return
	}
	todoToUpdate := m.displayedCompleted[m.cursor]
	for i := range m.completed {
		if m.completed[i].Text == todoToUpdate.Text && m.completed[i].CreatedAt.Equal(todoToUpdate.CreatedAt) {
			updateFn(&m.completed[i])
			break
		}
	}
	m.updateDisplayedCompleted()
}

// swapTodos swaps two adjacent items in a list and saves
func swapTodos(list []Todo, idx1, idx2 int, filename string) {
	list[idx1], list[idx2] = list[idx2], list[idx1]
	saveTodos(filename, list)
}

func initialModel() model {
	m := model{
		backlog:     loadTodos(backlogFile),
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
					if m.currentView == viewBacklog {
						m.backlog = append(m.backlog, newTodo)
						saveTodos(backlogFile, m.backlog)
					} else {
						m.inProgress = append(m.inProgress, newTodo)
						saveTodos(inProgressFile, m.inProgress)
					}
					m.message = "Todo added!"
				}
				m.adding = false
				m.newTodo = ""
			case "esc":
				m.adding = false
				m.newTodo = ""
				m.message = "Cancelled"
			default:
				handleTextInput(msg.String(), &m.newTodo)
			}
			return m, nil
		}

		if m.editingDescription {
			switch msg.String() {
			case "enter":
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					switch m.currentView {
					case viewBacklog:
						m.backlog[m.cursor].Description = m.newDescription
						saveTodos(backlogFile, m.backlog)
					case viewInProgress:
						m.inProgress[m.cursor].Description = m.newDescription
						saveTodos(inProgressFile, m.inProgress)
					case viewCompleted:
						m.updateCompletedTodo(func(t *Todo) {
							t.Description = m.newDescription
						})
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
			default:
				handleTextInput(msg.String(), &m.newDescription)
			}
			return m, nil
		}

		if m.renamingTodo {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.newTodoName) != "" {
					currentList := m.getCurrentList()
					if len(currentList) > 0 && m.cursor < len(currentList) {
						switch m.currentView {
						case viewBacklog:
							m.backlog[m.cursor].Text = m.newTodoName
							saveTodos(backlogFile, m.backlog)
						case viewInProgress:
							m.inProgress[m.cursor].Text = m.newTodoName
							saveTodos(inProgressFile, m.inProgress)
						case viewCompleted:
							m.updateCompletedTodo(func(t *Todo) {
								t.Text = m.newTodoName
							})
							saveTodos(completedFile, m.completed)
						}
						m.message = "Todo renamed!"
					}
				}
				m.renamingTodo = false
				m.newTodoName = ""
			case "esc":
				m.renamingTodo = false
				m.newTodoName = ""
				m.message = "Cancelled"
			default:
				handleTextInput(msg.String(), &m.newTodoName)
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
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor < len(m.backlog)-1 {
				swapTodos(m.backlog, m.cursor, m.cursor+1, backlogFile)
				m.cursor++
				m.message = "Todo moved down"
			} else if m.currentView == viewInProgress && len(m.inProgress) > 0 && m.cursor < len(m.inProgress)-1 {
				swapTodos(m.inProgress, m.cursor, m.cursor+1, inProgressFile)
				m.cursor++
				m.message = "Todo moved down"
			}

		case "K":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor > 0 {
				swapTodos(m.backlog, m.cursor, m.cursor-1, backlogFile)
				m.cursor--
				m.message = "Todo moved up"
			} else if m.currentView == viewInProgress && len(m.inProgress) > 0 && m.cursor > 0 {
				swapTodos(m.inProgress, m.cursor, m.cursor-1, inProgressFile)
				m.cursor--
				m.message = "Todo moved up"
			}

		case "h":
			switch m.currentView {
			case viewInProgress:
				m.currentView = viewBacklog
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			case viewCompleted:
				m.currentView = viewInProgress
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			}

		case "l":
			switch m.currentView {
			case viewBacklog:
				m.currentView = viewInProgress
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			case viewInProgress:
				m.currentView = viewCompleted
				m.updateDisplayedCompleted()
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			}

		case "a":
			if m.currentView == viewBacklog || m.currentView == viewInProgress {
				m.adding = true
				m.newTodo = ""
				m.message = ""
			}

		case "d":
			// Check if there's a todo to delete
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				m.confirmingDelete = true
				m.message = ""
			}

		case "y":
			if m.confirmingDelete {
				// Proceed with deletion
				switch m.currentView {
				case viewBacklog:
					if len(m.backlog) > 0 && m.cursor < len(m.backlog) {
						m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
						if m.cursor >= len(m.backlog) && m.cursor > 0 {
							m.cursor--
						}
						saveTodos(backlogFile, m.backlog)
						m.message = "Todo deleted"
					}
				case viewInProgress:
					if len(m.inProgress) > 0 && m.cursor < len(m.inProgress) {
						m.inProgress = append(m.inProgress[:m.cursor], m.inProgress[m.cursor+1:]...)
						if m.cursor >= len(m.inProgress) && m.cursor > 0 {
							m.cursor--
						}
						saveTodos(inProgressFile, m.inProgress)
						m.message = "Todo deleted"
					}
				case viewCompleted:
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
				m.confirmingDelete = false
			}

		case "n":
			if m.confirmingDelete {
				m.confirmingDelete = false
				m.message = "Deletion cancelled"
			} else {
				// Rename mode
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					m.renamingTodo = true
					m.newTodoName = currentList[m.cursor].Text
					m.message = ""
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

		case "r":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor < len(m.backlog) {
				todo := m.backlog[m.cursor]
				m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
				m.inProgress = append(m.inProgress, todo)
				if m.cursor >= len(m.backlog) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(backlogFile, m.backlog)
				saveTodos(inProgressFile, m.inProgress)
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
	switch m.currentView {
	case viewBacklog:
		return m.backlog
	case viewInProgress:
		return m.inProgress
	default:
		return m.displayedCompleted
	}
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

	switch m.currentView {
	case viewBacklog:
		s.WriteString("  [BACKLOG]  In Progress  Completed\n\n")
	case viewInProgress:
		s.WriteString("  Backlog  [IN PROGRESS]  Completed\n\n")
	case viewCompleted:
		s.WriteString("  Backlog  In Progress  [COMPLETED]\n\n")
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
	} else if m.renamingTodo {
		s.WriteString("  Rename todo: " + m.newTodoName + "_\n")
		s.WriteString("  (press Enter to save, Esc to cancel)\n\n")
	} else if m.confirmingDelete {
		s.WriteString("  Are you sure you want to delete this todo? (y/n)\n\n")
	} else {
		s.WriteString("  Commands:\n")
		s.WriteString("  j/k: move down/up  J/K: reorder (backlog/in progress)  h/l: switch views\n")
		s.WriteString("  a: add  d: delete  x: mark complete  u: undo complete  r: move to in progress\n")
		s.WriteString("  i: toggle description  I: toggle all descriptions  e: edit description  n: rename  q: quit\n\n")
	}

	if m.message != "" {
		s.WriteString("  " + m.message + "\n")
	}

	return s.String()
}
