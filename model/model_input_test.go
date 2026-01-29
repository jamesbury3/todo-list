package model

import "testing"

func TestIsSpecialKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"ctrl+c", "ctrl+c", true},
		{"up arrow", "up", true},
		{"down arrow", "down", true},
		{"left arrow", "left", false},   // Now handled by text input
		{"right arrow", "right", false}, // Now handled by text input
		{"ctrl+z", "ctrl+z", true},
		{"ctrl+a", "ctrl+a", false}, // Now handled by text input (home)
		{"ctrl+e", "ctrl+e", false}, // Now handled by text input (end)
		{"alt+enter", "alt+enter", true},
		{"f1", "f1", true},
		{"f12", "f12", true},
		{"ctrl+anything", "ctrl+x", true},
		{"alt+anything", "alt+x", true},
		{"meta+anything", "meta+x", true},
		{"regular character", "a", false},
		{"number", "5", false},
		{"space", " ", false},
		{"enter", "enter", false},
		{"backspace", "backspace", false},
		{"esc", "esc", false},
		{"home", "home", false},     // Now handled by text input
		{"end", "end", false},       // Now handled by text input
		{"delete", "delete", false}, // Now handled by text input
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSpecialKey(tt.key)
			if result != tt.expected {
				t.Errorf("isSpecialKey(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestHandleTextInput(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		initialText    string
		initialCursor  int
		expectedText   string
		expectedCursor int
	}{
		{"add character at end", "a", "hello", 5, "helloa", 6},
		{"add character in middle", "x", "hello", 2, "hexllo", 3},
		{"add multiple characters at end", "world", "hello ", 6, "hello world", 11},
		{"backspace at end", "backspace", "hello", 5, "hell", 4},
		{"backspace in middle", "backspace", "hello", 3, "helo", 2},
		{"backspace on empty", "backspace", "", 0, "", 0},
		{"backspace with UTF-8", "backspace", "hello世", 6, "hello", 5},
		{"strip bracketed paste prefix", "[abc", "test", 4, "testabc", 7},
		{"strip bracketed paste suffix", "abc]", "test", 4, "testabc", 7},
		{"ignore special key up", "up", "hello", 5, "hello", 5},
		{"move cursor left", "left", "hello", 5, "hello", 4},
		{"move cursor right", "right", "hello", 2, "hello", 3},
		{"move cursor left at start", "left", "hello", 0, "hello", 0},
		{"move cursor right at end", "right", "hello", 5, "hello", 5},
		{"home key", "home", "hello", 3, "hello", 0},
		{"end key", "end", "hello", 2, "hello", 5},
		{"ctrl+a (home)", "ctrl+a", "hello", 3, "hello", 0},
		{"ctrl+e (end)", "ctrl+e", "hello", 2, "hello", 5},
		{"delete at end", "delete", "hello", 5, "hello", 5},
		{"delete in middle", "delete", "hello", 2, "helo", 2},
		{"add space", " ", "hello", 5, "hello ", 6},
		{"add number", "5", "task", 4, "task5", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.initialText
			cursor := tt.initialCursor
			handleTextInput(tt.key, &text, &cursor)
			if text != tt.expectedText {
				t.Errorf("handleTextInput(%q, %q, %d) text = %q, want %q", tt.key, tt.initialText, tt.initialCursor, text, tt.expectedText)
			}
			if cursor != tt.expectedCursor {
				t.Errorf("handleTextInput(%q, %q, %d) cursor = %d, want %d", tt.key, tt.initialText, tt.initialCursor, cursor, tt.expectedCursor)
			}
		})
	}
}
