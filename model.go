package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
					m.newTodo = m.newTodo[:len(m.newTodo)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.newTodo += msg.String()
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

		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.message = ""

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
			}

		case "l":
			if m.currentView == viewInProgress {
				m.currentView = viewCompleted
				m.updateDisplayedCompleted()
				m.cursor = 0
				m.message = ""
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
	var s strings.Builder

	s.WriteString("\n  📝 TODO LIST\n\n")

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

			// Format the display based on view
			if m.currentView == viewCompleted && todo.CompletedAt != nil {
				completedTime := todo.CompletedAt.Format("Jan 2, 15:04")
				s.WriteString(fmt.Sprintf("  %s %s [%s]\n", cursor, todo.Text, completedTime))
			} else {
				createdTime := todo.CreatedAt.Format("Jan 2, 15:04")
				s.WriteString(fmt.Sprintf("  %s %s [%s]\n", cursor, todo.Text, createdTime))
			}
		}
	}

	s.WriteString("\n")

	if m.adding {
		s.WriteString("  Add new todo: " + m.newTodo + "_\n")
		s.WriteString("  (press Enter to save, Esc to cancel)\n\n")
	} else {
		s.WriteString("  Commands:\n")
		s.WriteString("  j/k: move down/up  J/K: reorder (in progress only)  h/l: switch views\n")
		s.WriteString("  a: add  d: delete  x: mark complete  u: undo complete  q: quit\n\n")
	}

	if m.message != "" {
		s.WriteString("  " + m.message + "\n")
	}

	return s.String()
}
