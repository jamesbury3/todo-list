package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Color styles
var (
	// View tabs
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Background(lipgloss.Color("236")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)

	// Cursor and items
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	todoTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	timestampStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true)

	// Headers and sections
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	countStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)

	// Input prompts
	promptStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	inputCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	helpTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)

	// Messages
	successMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	errorMessageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	infoMessageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))

	// Commands
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
)

func isSpecialKey(key string) bool {
	// Filter out special keys and control sequences that shouldn't be added to text input
	// Note: left, right, home, end, delete are handled by handleTextInput and are not filtered
	specialKeys := []string{
		"ctrl+c", "ctrl+d", "ctrl+z", "ctrl+k", "ctrl+u",
		"up", "down",
		"pgup", "pgdown",
		"insert",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
		"alt+enter", "shift+enter",
	}

	for _, sk := range specialKeys {
		if key == sk {
			return true
		}
	}

	// Filter out other control sequences (keys starting with ctrl+, alt+, etc.)
	// except ctrl+a and ctrl+e which are handled for home/end navigation
	if key == "ctrl+a" || key == "ctrl+e" {
		return false
	}
	if strings.HasPrefix(key, "ctrl+") || strings.HasPrefix(key, "alt+") || strings.HasPrefix(key, "meta+") {
		return true
	}

	return false
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

// handleTextInput processes keyboard input for text entry fields with cursor support
func handleTextInput(key string, currentText *string, cursorPos *int) bool {
	runes := []rune(*currentText)
	textLen := len(runes)

	// Ensure cursor position is within bounds
	if *cursorPos < 0 {
		*cursorPos = 0
	}
	if *cursorPos > textLen {
		*cursorPos = textLen
	}

	switch key {
	case "left":
		if *cursorPos > 0 {
			*cursorPos--
		}
		return true
	case "right":
		if *cursorPos < textLen {
			*cursorPos++
		}
		return true
	case "home", "ctrl+a":
		*cursorPos = 0
		return true
	case "end", "ctrl+e":
		*cursorPos = textLen
		return true
	case "backspace":
		if *cursorPos > 0 {
			runes = append(runes[:*cursorPos-1], runes[*cursorPos:]...)
			*currentText = string(runes)
			*cursorPos--
		}
		return true
	case "delete":
		if *cursorPos < textLen {
			runes = append(runes[:*cursorPos], runes[*cursorPos+1:]...)
			*currentText = string(runes)
		}
		return true
	default:
		if !isSpecialKey(key) {
			// Strip bracketed paste markers if present
			key = strings.TrimPrefix(key, "[")
			key = strings.TrimSuffix(key, "]")
			// Insert at cursor position
			keyRunes := []rune(key)
			runes = append(runes[:*cursorPos], append(keyRunes, runes[*cursorPos:]...)...)
			*currentText = string(runes)
			*cursorPos += len(keyRunes)
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
		ready:       loadTodos(readyFile),
		completed:   loadTodos(completedFile),
		cursor:      0,
		currentView: viewReady,
	}
	m.updateDisplayedCompleted()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
						saveTodos(backlogFile, m.backlog)
					} else {
						m.ready = append(m.ready, newTodo)
						saveTodos(readyFile, m.ready)
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
					switch m.currentView {
					case viewBacklog:
						m.backlog[m.cursor].Description = m.newDescription
						saveTodos(backlogFile, m.backlog)
					case viewReady:
						m.ready[m.cursor].Description = m.newDescription
						saveTodos(readyFile, m.ready)
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
							saveTodos(backlogFile, m.backlog)
						case viewReady:
							m.ready[m.cursor].Text = capitalizedName
							saveTodos(readyFile, m.ready)
						case viewCompleted:
							m.updateCompletedTodo(func(t *Todo) {
								t.Text = capitalizedName
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
				handleTextInput(msg.String(), &m.newTodoName, &m.textInputCursor)
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
			} else if m.currentView == viewReady && len(m.ready) > 0 && m.cursor < len(m.ready)-1 {
				swapTodos(m.ready, m.cursor, m.cursor+1, readyFile)
				m.cursor++
				m.message = "Todo moved down"
			}

		case "K":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor > 0 {
				swapTodos(m.backlog, m.cursor, m.cursor-1, backlogFile)
				m.cursor--
				m.message = "Todo moved up"
			} else if m.currentView == viewReady && len(m.ready) > 0 && m.cursor > 0 {
				swapTodos(m.ready, m.cursor, m.cursor-1, readyFile)
				m.cursor--
				m.message = "Todo moved up"
			}

		case "h":
			switch m.currentView {
			case viewReady:
				m.currentView = viewBacklog
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			case viewCompleted:
				m.currentView = viewReady
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			}

		case "l":
			switch m.currentView {
			case viewBacklog:
				m.currentView = viewReady
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
			case viewReady:
				m.currentView = viewCompleted
				m.updateDisplayedCompleted()
				m.cursor = 0
				m.message = ""
				m.showingDescription = false
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
				case viewReady:
					if len(m.ready) > 0 && m.cursor < len(m.ready) {
						m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
						if m.cursor >= len(m.ready) && m.cursor > 0 {
							m.cursor--
						}
						saveTodos(readyFile, m.ready)
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
					m.textInputCursor = len([]rune(m.newTodoName))
					m.message = ""
				}
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
				saveTodos(readyFile, m.ready)
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
						m.ready = append(m.ready, todoToUndo)
						break
					}
				}
				m.updateDisplayedCompleted()
				if m.cursor >= len(m.displayedCompleted) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(readyFile, m.ready)
				saveTodos(completedFile, m.completed)
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
					saveTodos(completedFile, m.completed)
					m.message = fmt.Sprintf("Backed up to %s and cleared completed todos!", backupFile)
				}
			}

		case "r":
			if m.currentView == viewBacklog && len(m.backlog) > 0 && m.cursor < len(m.backlog) {
				todo := m.backlog[m.cursor]
				m.backlog = append(m.backlog[:m.cursor], m.backlog[m.cursor+1:]...)
				m.ready = append(m.ready, todo)
				if m.cursor >= len(m.backlog) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(backlogFile, m.backlog)
				saveTodos(readyFile, m.ready)
				m.message = "Todo moved to ready!"
			}

		case "b":
			if m.currentView == viewReady && len(m.ready) > 0 && m.cursor < len(m.ready) {
				todo := m.ready[m.cursor]
				m.ready = append(m.ready[:m.cursor], m.ready[m.cursor+1:]...)
				m.backlog = append(m.backlog, todo)
				if m.cursor >= len(m.ready) && m.cursor > 0 {
					m.cursor--
				}
				saveTodos(readyFile, m.ready)
				saveTodos(backlogFile, m.backlog)
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
				m.editingDescription = true
				m.newDescription = currentList[m.cursor].Description
				m.textInputCursor = len([]rune(m.newDescription))
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
	case viewReady:
		return m.ready
	default:
		return m.displayedCompleted
	}
}

// renderTextWithCursor inserts a cursor indicator at the specified position
func renderTextWithCursor(text string, cursorPos int) string {
	runes := []rune(text)
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}
	return string(runes[:cursorPos]) + "│" + string(runes[cursorPos:])
}

// renderColoredTextWithCursor renders text with a colored cursor at the specified position
func renderColoredTextWithCursor(text string, cursorPos int) string {
	runes := []rune(text)
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}
	beforeCursor := todoTextStyle.Render(string(runes[:cursorPos]))
	cursor := inputCursorStyle.Render("│")
	afterCursor := todoTextStyle.Render(string(runes[cursorPos:]))
	return beforeCursor + cursor + afterCursor
}

// renderWrappedTextWithCursor renders wrapped text with a cursor, returning multiple lines
func renderWrappedTextWithCursor(text string, cursorPos int, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{renderColoredTextWithCursor(text, cursorPos)}
	}

	runes := []rune(text)
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Wrap the text first
	wrappedLines := wrapText(text, maxWidth)
	if len(wrappedLines) == 0 {
		return []string{inputCursorStyle.Render("│")}
	}

	// Find which line contains the cursor
	var result []string
	charCount := 0
	for _, line := range wrappedLines {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)

		if cursorPos >= charCount && cursorPos <= charCount+lineLen {
			// Cursor is in this line
			localCursorPos := cursorPos - charCount
			beforeCursor := todoTextStyle.Render(string(lineRunes[:localCursorPos]))
			cursor := inputCursorStyle.Render("│")
			afterCursor := todoTextStyle.Render(string(lineRunes[localCursorPos:]))
			result = append(result, beforeCursor+cursor+afterCursor)
		} else {
			// No cursor in this line
			result = append(result, todoTextStyle.Render(line))
		}

		charCount += lineLen
	}

	return result
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

// countCompletedToday returns the number of todos completed today
func (m *model) countCompletedToday() int {
	today := time.Now()
	todayYear, todayMonth, todayDay := today.Date()
	count := 0

	for _, todo := range m.completed {
		if todo.CompletedAt != nil {
			completedYear, completedMonth, completedDay := todo.CompletedAt.Date()
			if todayYear == completedYear && todayMonth == completedMonth && todayDay == completedDay {
				count++
			}
		}
	}

	return count
}

// truncateString truncates a string to maxLen, accounting for ANSI escape codes
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	// Count visible runes (excluding ANSI codes)
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// wrapText wraps text to fit within maxWidth, breaking at word boundaries when possible
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	runes := []rune(text)
	if len(runes) <= maxWidth {
		return []string{text}
	}

	var lines []string
	var currentLine []rune
	var currentWord []rune

	for i, r := range runes {
		currentWord = append(currentWord, r)

		// Check if we hit a space or end of text
		isSpace := r == ' ' || r == '\t'
		isEnd := i == len(runes)-1

		if isSpace || isEnd {
			// Would adding this word exceed the width?
			testLine := append(currentLine, currentWord...)

			if len(testLine) > maxWidth && len(currentLine) > 0 {
				// Save current line and start new one with the word
				lines = append(lines, string(currentLine))
				currentLine = currentWord
			} else {
				// Add word to current line
				currentLine = testLine
			}

			// Reset word buffer
			currentWord = []rune{}
		}
	}

	// Add any remaining content
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}

	return lines
}

func (m model) View() string {
	s := strings.Builder{}

	// Use alternate screen buffer approach: render content without leading newline
	// to prevent shifting when terminal resizes

	// Calculate available width for content (accounting for padding and margins)
	availableWidth := m.width
	if availableWidth <= 0 {
		availableWidth = 80 // Default width if not set yet
	}
	// Reserve space for padding, cursor, etc. (roughly 10 chars per line)
	maxTextWidth := availableWidth - 35 // Account for "  > ", timestamp, indicators

	// Render view tabs with colors
	backlogTab := "BACKLOG"
	readyTab := "READY"
	completedTab := "COMPLETED"

	switch m.currentView {
	case viewBacklog:
		s.WriteString("  " + activeTabStyle.Render(backlogTab) + "  " + inactiveTabStyle.Render(readyTab) + "  " + inactiveTabStyle.Render(completedTab) + "\n\n")
	case viewReady:
		s.WriteString("  " + inactiveTabStyle.Render(backlogTab) + "  " + activeTabStyle.Render(readyTab) + "  " + inactiveTabStyle.Render(completedTab) + "\n\n")
	case viewCompleted:
		s.WriteString("  " + inactiveTabStyle.Render(backlogTab) + "  " + inactiveTabStyle.Render(readyTab) + "  " + activeTabStyle.Render(completedTab) + "\n\n")
	}

	// Display count of todos completed today
	completedToday := m.countCompletedToday()
	s.WriteString("  " + headerStyle.Render("Completed today:") + " " + countStyle.Render(fmt.Sprintf("%d", completedToday)) + "\n\n")

	currentList := m.getCurrentList()

	if len(currentList) == 0 {
		s.WriteString("  " + infoMessageStyle.Render("No todos") + "\n")
	} else {
		for i, todo := range currentList {
			cursor := " "
			if i == m.cursor {
				cursor = cursorStyle.Render(">")
			}

			// Add description indicator
			indicator := ""
			if todo.Description != "" {
				indicator = " 📄"
			}

			// Wrap todo text if needed
			wrappedLines := wrapText(todo.Text, maxTextWidth)

			// Format the display based on view
			var timestamp string
			if m.currentView == viewCompleted && todo.CompletedAt != nil {
				completedTime := todo.CompletedAt.Format("Jan 2, 15:04")
				timestamp = timestampStyle.Render("[" + completedTime + "]")
			} else {
				createdTime := todo.CreatedAt.Format("Jan 2, 15:04")
				timestamp = timestampStyle.Render("[" + createdTime + "]")
			}

			// Render first line with cursor and timestamp
			if len(wrappedLines) > 0 {
				firstLine := wrappedLines[0]
				// Add indicator to the first line
				if len(wrappedLines) == 1 {
					firstLine += indicator
				}
				todoText := todoTextStyle.Render(firstLine)
				s.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, todoText, timestamp))

				// Render additional wrapped lines with proper indentation
				for j := 1; j < len(wrappedLines); j++ {
					line := wrappedLines[j]
					// Add indicator to the last line
					if j == len(wrappedLines)-1 {
						line += indicator
					}
					todoText := todoTextStyle.Render(line)
					s.WriteString(fmt.Sprintf("     %s\n", todoText))
				}
			}

			// Show description if toggled and cursor is on this todo, or if showing all descriptions
			if todo.Description != "" && ((m.showingDescription && i == m.cursor) || m.showingAllDescriptions) {
				// Wrap description text
				descLines := wrapText(todo.Description, maxTextWidth-5)
				for j, descLine := range descLines {
					if j == 0 {
						s.WriteString("     " + descriptionStyle.Render("└─ "+descLine) + "\n")
					} else {
						s.WriteString("     " + descriptionStyle.Render("   "+descLine) + "\n")
					}
				}
			}
		}
	}

	s.WriteString("\n")

	if m.adding {
		// Wrap input text display if too wide
		inputMaxWidth := maxTextWidth + 10 // Slightly more space for input
		wrappedLines := renderWrappedTextWithCursor(m.newTodo, m.textInputCursor, inputMaxWidth)
		s.WriteString("  " + promptStyle.Render("Add new todo:") + " " + wrappedLines[0] + "\n")
		for i := 1; i < len(wrappedLines); i++ {
			s.WriteString("                " + wrappedLines[i] + "\n")
		}
		s.WriteString("  " + helpTextStyle.Render("(press Enter to save, Esc to cancel, arrows to navigate)") + "\n\n")
	} else if m.editingDescription {
		// Wrap input text display if too wide
		inputMaxWidth := maxTextWidth + 10
		wrappedLines := renderWrappedTextWithCursor(m.newDescription, m.textInputCursor, inputMaxWidth)
		s.WriteString("  " + promptStyle.Render("Edit description:") + " " + wrappedLines[0] + "\n")
		for i := 1; i < len(wrappedLines); i++ {
			s.WriteString("                   " + wrappedLines[i] + "\n")
		}
		s.WriteString("  " + helpTextStyle.Render("(press Enter to save, Esc to cancel, arrows to navigate)") + "\n\n")
	} else if m.renamingTodo {
		// Wrap input text display if too wide
		inputMaxWidth := maxTextWidth + 10
		wrappedLines := renderWrappedTextWithCursor(m.newTodoName, m.textInputCursor, inputMaxWidth)
		s.WriteString("  " + promptStyle.Render("Rename todo:") + " " + wrappedLines[0] + "\n")
		for i := 1; i < len(wrappedLines); i++ {
			s.WriteString("              " + wrappedLines[i] + "\n")
		}
		s.WriteString("  " + helpTextStyle.Render("(press Enter to save, Esc to cancel, arrows to navigate)") + "\n\n")
	} else if m.confirmingDelete {
		s.WriteString("  " + errorMessageStyle.Render("Are you sure you want to delete this todo? (y/n)") + "\n\n")
	} else {
		s.WriteString("  " + headerStyle.Render("Commands:") + "\n")
		s.WriteString("  " + commandStyle.Render("j/k: move down/up  J/K: reorder (backlog/ready)  h/l: switch views") + "\n")
		if m.currentView == viewCompleted {
			s.WriteString("  " + commandStyle.Render("d: delete  u: undo complete  B: backup and clear completed") + "\n")
		} else if m.currentView == viewReady {
			s.WriteString("  " + commandStyle.Render("a: add  d: delete  x: mark complete  b: move to backlog") + "\n")
		} else if m.currentView == viewBacklog {
			s.WriteString("  " + commandStyle.Render("a: add  d: delete  r: move to ready") + "\n")
		} else {
			log.Fatalf("Invalid view: %v", m.currentView)
		}
		s.WriteString("  " + commandStyle.Render("i: toggle description  I: toggle all descriptions  e: edit description  n: rename  q: quit") + "\n\n")
	}

	if m.message != "" {
		// Determine message style based on content
		msgStyle := infoMessageStyle
		if strings.Contains(strings.ToLower(m.message), "added") ||
			strings.Contains(strings.ToLower(m.message), "completed") ||
			strings.Contains(strings.ToLower(m.message), "moved") ||
			strings.Contains(strings.ToLower(m.message), "updated") ||
			strings.Contains(strings.ToLower(m.message), "renamed") ||
			strings.Contains(strings.ToLower(m.message), "backed up") {
			msgStyle = successMessageStyle
		} else if strings.Contains(strings.ToLower(m.message), "cancelled") ||
			strings.Contains(strings.ToLower(m.message), "failed") ||
			strings.Contains(strings.ToLower(m.message), "error") {
			msgStyle = errorMessageStyle
		}
		s.WriteString("  " + msgStyle.Render(m.message) + "\n")
	}

	return s.String()
}
