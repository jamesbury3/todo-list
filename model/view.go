package model

import (
	"fmt"
	"log"
	"strings"
)

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

// View renders the model's UI
func (m Model) View() string {
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
			if len(todo.Description) > 0 {
				indicator = fmt.Sprintf(" 📄×%d", len(todo.Description))
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

			// Show descriptions if toggled and cursor is on this todo, or if showing all descriptions
			if len(todo.Description) > 0 && ((m.showingDescription && i == m.cursor) || m.showingAllDescriptions) {
				for descIdx, desc := range todo.Description {
					// Wrap description text
					descLines := wrapText(desc, maxTextWidth-5)

					// Add cursor indicator if in navigation mode
					descCursor := ""
					if m.navigatingDescriptions && i == m.cursor && descIdx == m.descriptionCursor {
						descCursor = cursorStyle.Render("►") + " "
					} else {
						descCursor = "  "
					}

					for j, descLine := range descLines {
						if j == 0 {
							s.WriteString("     " + descCursor + descriptionStyle.Render("└─ "+descLine) + "\n")
						} else {
							s.WriteString("     " + "   " + descriptionStyle.Render("   "+descLine) + "\n")
						}
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
	} else if m.confirmingDeleteDesc {
		s.WriteString("  " + errorMessageStyle.Render("Are you sure you want to delete this description? (y/n)") + "\n\n")
	} else if m.showingCommands {
		s.WriteString("  " + headerStyle.Render("Commands:") + "\n")
		s.WriteString("  " + commandStyle.Render("j/k: move down/up  g/G: go to top/bottom  J/K: reorder (backlog/ready)  h/l: switch views") + "\n")
		if m.currentView == viewCompleted {
			s.WriteString("  " + commandStyle.Render("d: delete  u: undo complete  B: backup and clear completed") + "\n")
		} else if m.currentView == viewReady {
			s.WriteString("  " + commandStyle.Render("a: add  d: delete  x: mark complete  b: move to backlog") + "\n")
		} else if m.currentView == viewBacklog {
			s.WriteString("  " + commandStyle.Render("a: add  d: delete  r: move to ready") + "\n")
		} else {
			log.Fatalf("Invalid view: %v", m.currentView)
		}
		s.WriteString("  " + commandStyle.Render("i: toggle description  I: toggle all descriptions  e: add description  enter: navigate descriptions  n: rename  ?: toggle help  q: quit") + "\n\n")
	} else {
		s.WriteString("  " + helpTextStyle.Render("Press ? for help") + "\n\n")
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
