package model

import (
	"testing"
	"time"
)

func TestViewRendering(t *testing.T) {
	now := time.Now()
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: now},
			{Text: "task2", Description: []string{"has description"}, CreatedAt: now},
		},
		cursor: 0,
	}

	view := m.View()

	// Check that the view contains expected elements
	if !contains(view, "BACKLOG") {
		t.Error("View should contain 'BACKLOG'")
	}
	if !contains(view, "task1") {
		t.Error("View should contain 'task1'")
	}
	if !contains(view, "task2") {
		t.Error("View should contain 'task2'")
	}
	if !contains(view, "📄") {
		t.Error("View should contain description indicator")
	}
	if !contains(view, "Press ? for help") {
		t.Error("View should contain help prompt")
	}
}

func TestViewRenderingEmptyList(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog:     []Todo{},
	}

	view := m.View()

	if !contains(view, "No todos") {
		t.Error("View should contain 'No todos' for empty list")
	}
}

func TestViewRenderingAddingMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		adding:      true,
		newTodo:     "new task",
	}

	view := m.View()

	if !contains(view, "Add new todo:") {
		t.Error("View should contain 'Add new todo:' prompt")
	}
	if !contains(view, "new task") {
		t.Error("View should contain the new todo text")
	}
	if !contains(view, "press Enter to save") {
		t.Error("View should contain save instruction")
	}
}

func TestViewRenderingDeleteConfirmation(t *testing.T) {
	m := Model{
		currentView:      viewBacklog,
		confirmingDelete: true,
	}

	view := m.View()

	if !contains(view, "Are you sure") {
		t.Error("View should contain delete confirmation prompt")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		maxWidth    int
		expected    []string
		description string
	}{
		{
			name:        "text shorter than maxWidth",
			text:        "short",
			maxWidth:    10,
			expected:    []string{"short"},
			description: "Short text should return single line",
		},
		{
			name:        "text equal to maxWidth",
			text:        "exactly10!",
			maxWidth:    10,
			expected:    []string{"exactly10!"},
			description: "Text exactly at maxWidth should return single line",
		},
		{
			name:        "text longer than maxWidth",
			text:        "This is a longer text that needs wrapping",
			maxWidth:    20,
			expected:    []string{"This is a longer ", "text that needs ", "wrapping"},
			description: "Long text should wrap at word boundaries",
		},
		{
			name:        "multiple words wrapping",
			text:        "one two three four five",
			maxWidth:    10,
			expected:    []string{"one two ", "three ", "four five"},
			description: "Multiple words should wrap correctly",
		},
		{
			name:        "long word exceeding maxWidth",
			text:        "supercalifragilisticexpialidocious",
			maxWidth:    10,
			expected:    []string{"supercalifragilisticexpialidocious"},
			description: "Single long word should not be split",
		},
		{
			name:        "empty string",
			text:        "",
			maxWidth:    10,
			expected:    []string{""},
			description: "Empty string should return single empty line",
		},
		{
			name:        "maxWidth zero",
			text:        "test",
			maxWidth:    0,
			expected:    []string{"test"},
			description: "MaxWidth <= 0 should return single line",
		},
		{
			name:        "maxWidth negative",
			text:        "test",
			maxWidth:    -5,
			expected:    []string{"test"},
			description: "Negative maxWidth should return single line",
		},
		{
			name:        "UTF-8 characters",
			text:        "世界 你好 测试 文本",
			maxWidth:    10,
			expected:    []string{"世界 你好 测试 ", "文本"},
			description: "UTF-8 characters should wrap correctly",
		},
		{
			name:        "text with spaces",
			text:        "a b c d e f g h",
			maxWidth:    5,
			expected:    []string{"a b ", "c d ", "e f ", "g h"},
			description: "Words separated by spaces should wrap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.maxWidth)

			if len(result) != len(tt.expected) {
				t.Errorf("wrapText() returned %d lines, want %d\nGot: %v\nWant: %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("wrapText() line[%d] = %q, want %q", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestRenderColoredTextWithCursor(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		cursorPos   int
		description string
		checkFunc   func(t *testing.T, result string)
	}{
		{
			name:        "cursor at start",
			text:        "hello",
			cursorPos:   0,
			description: "Cursor at position 0",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hello") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:        "cursor in middle",
			text:        "hello",
			cursorPos:   2,
			description: "Cursor at position 2",
			checkFunc: func(t *testing.T, result string) {
				// Result contains styled text with ANSI codes, just verify it's not empty
				if result == "" {
					t.Error("Result should not be empty")
				}
			},
		},
		{
			name:        "cursor at end",
			text:        "hello",
			cursorPos:   5,
			description: "Cursor at end of text",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hello") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:        "cursor beyond text",
			text:        "hi",
			cursorPos:   10,
			description: "Cursor position beyond text length",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hi") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:        "negative cursor",
			text:        "test",
			cursorPos:   -1,
			description: "Negative cursor position",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "test") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:        "empty text",
			text:        "",
			cursorPos:   0,
			description: "Empty text with cursor",
			checkFunc: func(t *testing.T, result string) {
				// Should just return cursor indicator
				if result == "" {
					t.Error("Result should not be empty")
				}
			},
		},
		{
			name:        "UTF-8 characters",
			text:        "世界",
			cursorPos:   1,
			description: "UTF-8 text with cursor",
			checkFunc: func(t *testing.T, result string) {
				// Result contains styled text with ANSI codes, just verify it's not empty
				if result == "" {
					t.Error("Result should not be empty for UTF-8 text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderColoredTextWithCursor(tt.text, tt.cursorPos)
			tt.checkFunc(t, result)
		})
	}
}

func TestInit(t *testing.T) {
	m := Model{}
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return tea.ClearScreen command")
	}
}
