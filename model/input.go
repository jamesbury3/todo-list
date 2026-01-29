package model

import "strings"

// isSpecialKey checks if a key is a special key that shouldn't be added to text input
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
