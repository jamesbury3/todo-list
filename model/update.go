package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all user input and state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.adding {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.newTodo) != "" {
					newTodo := Todo{
						Text:      capitalizeFirst(m.newTodo),
						CreatedAt: time.Now(),
					}
					if m.currentView == viewBacklog {
						m.backlog = append(m.backlog, newTodo)
						if cmd := m.save(backlogFile, m.backlog); cmd != nil {
							return m, cmd
						}
					} else {
						m.ready = append(m.ready, newTodo)
						if cmd := m.save(readyFile, m.ready); cmd != nil {
							return m, cmd
						}
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
				handleTextInput(msg.String(), &m.newTodo, &m.textInputCursor)
			}
			return m, nil
		}

		if m.editingDescription {
			switch msg.String() {
			case "enter":
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					trimmedDesc := strings.TrimSpace(m.newDescription)
					if trimmedDesc != "" {
						switch m.currentView {
						case viewBacklog:
							if m.navigatingDescriptions && m.descriptionCursor < len(m.backlog[m.cursor].Description) {
								// Update existing description
								m.backlog[m.cursor].Description[m.descriptionCursor] = trimmedDesc
							} else {
								// Prepend new description
								m.backlog[m.cursor].Description = append([]string{trimmedDesc}, m.backlog[m.cursor].Description...)
							}
							if cmd := m.save(backlogFile, m.backlog); cmd != nil {
								return m, cmd
							}
						case viewReady:
							if m.navigatingDescriptions && m.descriptionCursor < len(m.ready[m.cursor].Description) {
								// Update existing description
								m.ready[m.cursor].Description[m.descriptionCursor] = trimmedDesc
							} else {
								// Prepend new description
								m.ready[m.cursor].Description = append([]string{trimmedDesc}, m.ready[m.cursor].Description...)
							}
							if cmd := m.save(readyFile, m.ready); cmd != nil {
								return m, cmd
							}
						case viewCompleted:
							m.updateCompletedTodo(func(t *Todo) {
								if m.navigatingDescriptions && m.descriptionCursor < len(t.Description) {
									// Update existing description
									t.Description[m.descriptionCursor] = trimmedDesc
								} else {
									// Prepend new description
									t.Description = append([]string{trimmedDesc}, t.Description...)
								}
							})
							if cmd := m.save(completedFile, m.completed); cmd != nil {
								return m, cmd
							}
						}
						m.message = "Description saved!"
						m.showingDescription = true
					}
				}
				m.editingDescription = false
				m.newDescription = ""
			case "esc":
				m.editingDescription = false
				m.newDescription = ""
				m.message = "Cancelled"
			default:
				handleTextInput(msg.String(), &m.newDescription, &m.textInputCursor)
			}
			return m, nil
		}

		if m.renamingTodo {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.newTodoName) != "" {
					currentList := m.getCurrentList()
					if len(currentList) > 0 && m.cursor < len(currentList) {
						capitalizedName := capitalizeFirst(m.newTodoName)
						switch m.currentView {
						case viewBacklog:
							m.backlog[m.cursor].Text = capitalizedName
							if cmd := m.save(backlogFile, m.backlog); cmd != nil {
								return m, cmd
							}
						case viewReady:
							m.ready[m.cursor].Text = capitalizedName
							if cmd := m.save(readyFile, m.ready); cmd != nil {
								return m, cmd
							}
						case viewCompleted:
							m.updateCompletedTodo(func(t *Todo) {
								t.Text = capitalizedName
							})
							if cmd := m.save(completedFile, m.completed); cmd != nil {
								return m, cmd
							}
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
				handleTextInput(msg.String(), &m.newTodoName, &m.textInputCursor)
			}
			return m, nil
		}

		// Handle description deletion confirmation (check this BEFORE navigatingDescriptions)
		if m.confirmingDeleteDesc {
			switch msg.String() {
			case "y":
				// Delete the description
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					switch m.currentView {
					case viewBacklog:
						descriptions := m.backlog[m.cursor].Description
						m.backlog[m.cursor].Description = append(descriptions[:m.descriptionCursor], descriptions[m.descriptionCursor+1:]...)
						if cmd := m.save(backlogFile, m.backlog); cmd != nil {
							return m, cmd
						}
					case viewReady:
						descriptions := m.ready[m.cursor].Description
						m.ready[m.cursor].Description = append(descriptions[:m.descriptionCursor], descriptions[m.descriptionCursor+1:]...)
						if cmd := m.save(readyFile, m.ready); cmd != nil {
							return m, cmd
						}
					case viewCompleted:
						m.updateCompletedTodo(func(t *Todo) {
							t.Description = append(t.Description[:m.descriptionCursor], t.Description[m.descriptionCursor+1:]...)
						})
						if cmd := m.save(completedFile, m.completed); cmd != nil {
							return m, cmd
						}
					}

					// Adjust cursor if needed
					currentList = m.getCurrentList()
					todo := currentList[m.cursor]
					if len(todo.Description) == 0 {
						// No more descriptions, exit navigation mode
						m.navigatingDescriptions = false
					} else if m.descriptionCursor >= len(todo.Description) {
						m.descriptionCursor = len(todo.Description) - 1
					}

					m.message = "Description deleted"
				}
				m.confirmingDeleteDesc = false
				return m, nil

			case "n", "esc":
				m.confirmingDeleteDesc = false
				m.message = "Deletion cancelled"
				return m, nil
			}
			return m, nil
		}

		// Handle todo deletion confirmation
		if m.confirmingDelete {
			switch msg.String() {
			case "y":
				// Proceed with deletion
				switch m.currentView {
				case viewBacklog:
					if len(m.backlog) > 0 && m.cursor < len(m.backlog) {
						m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
						if m.cursor >= len(m.backlog) && m.cursor > 0 {
							m.cursor--
						}
						if cmd := m.save(backlogFile, m.backlog); cmd != nil {
							return m, cmd
						}
						m.message = "Todo deleted"
					}
				case viewReady:
					if len(m.ready) > 0 && m.cursor < len(m.ready) {
						m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
						if m.cursor >= len(m.ready) && m.cursor > 0 {
							m.cursor--
						}
						if cmd := m.save(readyFile, m.ready); cmd != nil {
							return m, cmd
						}
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
						if cmd := m.save(completedFile, m.completed); cmd != nil {
							return m, cmd
						}
						m.message = "Todo deleted"
					}
				}
				m.confirmingDelete = false
				return m, nil

			case "n", "esc":
				m.confirmingDelete = false
				m.message = "Deletion cancelled"
				return m, nil
			}
			return m, nil
		}
		// Handle description navigation mode
		if m.navigatingDescriptions {
			switch msg.String() {
			case "j":
				currentList := m.getCurrentList()
				if len(currentList) > 0 && m.cursor < len(currentList) {
					todo := currentList[m.cursor]
					if m.descriptionCursor < len(todo.Description)-1 {
						m.descriptionCursor++
					}
				}
				m.message = ""
				return m, nil

			case "k":
				if m.descriptionCursor > 0 {
					m.descriptionCursor--
				}
				m.message = ""
				return m, nil

			case "d":
				// Initiate description deletion
				m.confirmingDeleteDesc = true
				m.message = ""
				return m, nil

			case "esc", "q":
				// Exit description navigation mode
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
				m.message = ""
				m.showingDescription = false
				return m, nil

			case "e":
				// Edit selected description (handled in main 'e' case below)

			default:
				// Ignore other keys in this mode
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				todo := currentList[m.cursor]
				if len(todo.Description) > 0 {
					// Enter description navigation mode
					m.navigatingDescriptions = true
					m.descriptionCursor = 0
					m.showingDescription = true
					m.message = "Description navigation mode (j/k to navigate, e to edit, d to delete, esc to exit)"
				} else {
					m.message = "No descriptions. Press 'e' to add one."
				}
			}

		case "j":
			currentList := m.getCurrentList()
			if m.cursor < len(currentList)-1 {
				m.cursor++
			}
			m.message = ""
			m.showingDescription = false
			// Exit description navigation when moving between todos
			m.navigatingDescriptions = false
			m.descriptionCursor = 0

		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.message = ""
			m.showingDescription = false
			// Exit description navigation when moving between todos
			m.navigatingDescriptions = false
			m.descriptionCursor = 0

		case "J":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor < len(m.backlog)-1 {
				swapTodos(m.backlog, m.cursor, m.cursor+1)
				if cmd := m.save(backlogFile, m.backlog); cmd != nil {
					return m, cmd
				}
				m.cursor++
				m.message = "Todo moved down"
			} else if m.currentView == viewReady && len(m.ready) > 0 && m.cursor < len(m.ready)-1 {
				swapTodos(m.ready, m.cursor, m.cursor+1)
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				m.cursor++
				m.message = "Todo moved down"
			}

		case "K":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor > 0 {
				swapTodos(m.backlog, m.cursor, m.cursor-1)
				if cmd := m.save(backlogFile, m.backlog); cmd != nil {
					return m, cmd
				}
				m.cursor--
				m.message = "Todo moved up"
			} else if m.currentView == viewReady && len(m.ready) > 0 && m.cursor > 0 {
				swapTodos(m.ready, m.cursor, m.cursor-1)
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				m.cursor--
				m.message = "Todo moved up"
			}

		case "t":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor > 0 {
				// Move current todo to the top
				todo := m.backlog[m.cursor]
				m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
				m.backlog = append([]Todo{todo}, m.backlog...)
				m.cursor = 0
				if cmd := m.save(backlogFile, m.backlog); cmd != nil {
					return m, cmd
				}
				m.message = "Todo moved to top"
			} else if m.currentView == viewReady && len(m.ready) > 0 && m.cursor > 0 {
				// Move current todo to the top
				todo := m.ready[m.cursor]
				m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
				m.ready = append([]Todo{todo}, m.ready...)
				m.cursor = 0
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				m.message = "Todo moved to top"
			}

		case "h":
			switch m.currentView {
			case viewReady:
				m.currentView = viewBacklog
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			case viewCompleted:
				m.currentView = viewReady
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			}

		case "l":
			switch m.currentView {
			case viewBacklog:
				m.currentView = viewReady
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			case viewReady:
				m.currentView = viewCompleted
				m.updateDisplayedCompleted()
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			}

		case "a":
			if m.currentView == viewBacklog || m.currentView == viewReady {
				m.adding = true
				m.newTodo = ""
				m.textInputCursor = 0
				m.message = ""
			}

		case "d":
			// Check if there's a todo to delete
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				m.confirmingDelete = true
				m.message = ""
			}

		case "n":
			// Rename mode
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList) {
				m.renamingTodo = true
				m.newTodoName = currentList[m.cursor].Text
				m.textInputCursor = len([]rune(m.newTodoName))
				m.message = ""
			}

		case "x":
			if m.currentView == viewReady && len(m.ready) > 0 && m.cursor < len(m.ready) {
				todo := m.ready[m.cursor]
				now := time.Now()
				todo.CompletedAt = &now
				m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
				m.completed = append(m.completed, todo)
				m.updateDisplayedCompleted()
				if m.cursor >= len(m.ready) && m.cursor > 0 {
					m.cursor--
				}
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				if cmd := m.save(completedFile, m.completed); cmd != nil {
					return m, cmd
				}
				m.message = "Todo completed!"
			}

		case "r":
			if m.currentView == viewCompleted && len(m.displayedCompleted) > 0 && m.cursor < len(m.displayedCompleted) {
				// Find and remove from the actual completed list
				todoToUndo := m.displayedCompleted[m.cursor]
				for i, todo := range m.completed {
					if todo.Text == todoToUndo.Text && todo.CreatedAt.Equal(todoToUndo.CreatedAt) {
						// Clear the completion timestamp
						todoToUndo.CompletedAt = nil
						m.completed = append(m.completed[:i], m.completed[i+1:]...)
						m.ready = append(m.ready, todoToUndo)
						break
					}
				}
				m.updateDisplayedCompleted()
				if m.cursor >= len(m.displayedCompleted) && m.cursor > 0 {
					m.cursor--
				}
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				if cmd := m.save(completedFile, m.completed); cmd != nil {
					return m, cmd
				}
				m.message = "Todo moved back to ready!"
			} else if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor < len(m.backlog) {
				todo := m.backlog[m.cursor]
				m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
				m.ready = append(m.ready, todo)
				if m.cursor >= len(m.backlog) && m.cursor > 0 {
					m.cursor--
				}
				if cmd := m.save(backlogFile, m.backlog); cmd != nil {
					return m, cmd
				}
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				m.message = "Todo moved to ready!"
			}

		case "B":
			if m.currentView == viewCompleted && len(m.completed) > 0 {
				// Backup all completed todos
				backupFile, err := backupCompletedTodos(m.completed)
				if err != nil {
					m.message = "Backup failed: " + err.Error()
				} else {
					// Clear the completed list
					m.completed = []Todo{}
					m.updateDisplayedCompleted()
					m.cursor = 0
					if cmd := m.save(completedFile, m.completed); cmd != nil {
						return m, cmd
					}
					m.message = fmt.Sprintf("Backed up to %s and cleared completed todos!", backupFile)
				}
			}

		case "b":
			if m.currentView == viewReady && len(m.ready) > 0 && m.cursor < len(m.ready) {
				todo := m.ready[m.cursor]
				m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
				m.backlog = append([]Todo{todo}, m.backlog...)
				if m.cursor >= len(m.ready) && m.cursor > 0 {
					m.cursor--
				}
				if cmd := m.save(readyFile, m.ready); cmd != nil {
					return m, cmd
				}
				if cmd := m.save(backlogFile, m.backlog); cmd != nil {
					return m, cmd
				}
				m.message = "Todo moved to backlog!"
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
				if m.navigatingDescriptions {
					// Editing existing description
					todo := currentList[m.cursor]
					if m.descriptionCursor < len(todo.Description) {
						m.editingDescription = true
						m.newDescription = todo.Description[m.descriptionCursor]
						m.textInputCursor = len([]rune(m.newDescription))
						m.message = ""
					}
				} else {
					// Always create NEW description
					m.editingDescription = true
					m.newDescription = ""
					m.textInputCursor = 0
					m.message = ""
				}
			}

		case "?":
			m.showingCommands = !m.showingCommands
			m.message = ""

		case "g":
			// Go to top
			m.cursor = 0
			m.message = ""
			m.showingDescription = false
			m.navigatingDescriptions = false
			m.descriptionCursor = 0

		case "G":
			// Go to bottom
			currentList := m.getCurrentList()
			if len(currentList) > 0 {
				m.cursor = len(currentList) - 1
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			}

		case "p":
			// Toggle prettify view (only in Completed tab)
			if m.currentView == viewCompleted {
				m.showingPrettify = !m.showingPrettify
				m.message = ""
				m.showingDescription = false
				m.navigatingDescriptions = false
				m.descriptionCursor = 0
			}

		case "P":
			// Export prettify view with backups to markdown file (only in Completed tab)
			if m.currentView == viewCompleted {
				allTodos := loadAllCompletedTodos()
				filename, err := exportMarkdownFile(allTodos, true)
				if err != nil {
					m.message = fmt.Sprintf("Failed to export markdown: %s", err.Error())
				} else {
					m.message = fmt.Sprintf("Exported to %s!", filename)
				}
			}

		case "esc":
			// Universal untoggle: hide all descriptions and exit pretty view
			if m.showingDescription || m.showingAllDescriptions || m.showingPrettify {
				m.showingDescription = false
				m.showingAllDescriptions = false
				m.showingPrettify = false
				m.message = ""
			}
		}
	}

	return m, nil
}
