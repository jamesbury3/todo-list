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
			{Text: "task2", Updates: []string{"has update"}, CreatedAt: now},
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
		t.Error("View should contain update indicator")
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
		name     string
		text     string
		maxWidth int
		expected []string
		update   string
	}{
		{
			name:     "text shorter than maxWidth",
			text:     "short",
			maxWidth: 10,
			expected: []string{"short"},
			update:   "Short text should return single line",
		},
		{
			name:     "text equal to maxWidth",
			text:     "exactly10!",
			maxWidth: 10,
			expected: []string{"exactly10!"},
			update:   "Text exactly at maxWidth should return single line",
		},
		{
			name:     "text longer than maxWidth",
			text:     "This is a longer text that needs wrapping",
			maxWidth: 20,
			expected: []string{"This is a longer ", "text that needs ", "wrapping"},
			update:   "Long text should wrap at word boundaries",
		},
		{
			name:     "multiple words wrapping",
			text:     "one two three four five",
			maxWidth: 10,
			expected: []string{"one two ", "three ", "four five"},
			update:   "Multiple words should wrap correctly",
		},
		{
			name:     "long word exceeding maxWidth",
			text:     "supercalifragilisticexpialidocious",
			maxWidth: 10,
			expected: []string{"supercalifragilisticexpialidocious"},
			update:   "Single long word should not be split",
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			expected: []string{""},
			update:   "Empty string should return single empty line",
		},
		{
			name:     "maxWidth zero",
			text:     "test",
			maxWidth: 0,
			expected: []string{"test"},
			update:   "MaxWidth <= 0 should return single line",
		},
		{
			name:     "maxWidth negative",
			text:     "test",
			maxWidth: -5,
			expected: []string{"test"},
			update:   "Negative maxWidth should return single line",
		},
		{
			name:     "UTF-8 characters",
			text:     "世界 你好 测试 文本",
			maxWidth: 10,
			expected: []string{"世界 你好 测试 ", "文本"},
			update:   "UTF-8 characters should wrap correctly",
		},
		{
			name:     "text with spaces",
			text:     "a b c d e f g h",
			maxWidth: 5,
			expected: []string{"a b ", "c d ", "e f ", "g h"},
			update:   "Words separated by spaces should wrap",
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
		name      string
		text      string
		cursorPos int
		update    string
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:      "cursor at start",
			text:      "hello",
			cursorPos: 0,
			update:    "Cursor at position 0",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hello") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:      "cursor in middle",
			text:      "hello",
			cursorPos: 2,
			update:    "Cursor at position 2",
			checkFunc: func(t *testing.T, result string) {
				// Result contains styled text with ANSI codes, just verify it's not empty
				if result == "" {
					t.Error("Result should not be empty")
				}
			},
		},
		{
			name:      "cursor at end",
			text:      "hello",
			cursorPos: 5,
			update:    "Cursor at end of text",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hello") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:      "cursor beyond text",
			text:      "hi",
			cursorPos: 10,
			update:    "Cursor position beyond text length",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "hi") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:      "negative cursor",
			text:      "test",
			cursorPos: -1,
			update:    "Negative cursor position",
			checkFunc: func(t *testing.T, result string) {
				if !contains(result, "test") {
					t.Error("Result should contain text")
				}
			},
		},
		{
			name:      "empty text",
			text:      "",
			cursorPos: 0,
			update:    "Empty text with cursor",
			checkFunc: func(t *testing.T, result string) {
				// Should just return cursor indicator
				if result == "" {
					t.Error("Result should not be empty")
				}
			},
		},
		{
			name:      "UTF-8 characters",
			text:      "世界",
			cursorPos: 1,
			update:    "UTF-8 text with cursor",
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

func TestRenderPrettifyView(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	completedNow := now
	completedYesterday := yesterday

	tests := []struct {
		name         string
		todos        []Todo
		title        string
		exitKey      string
		wantContains []string
		update       string
	}{
		{
			name:    "empty todos",
			todos:   []Todo{},
			title:   "TEST VIEW",
			exitKey: "p",
			wantContains: []string{
				"TEST VIEW",
				"No completed todos",
				"Press p to exit prettify view",
			},
			update: "Should show empty message",
		},
		{
			name: "single todo",
			todos: []Todo{
				{Text: "Test task", CreatedAt: now, CompletedAt: &completedNow},
			},
			title:   "COMPLETED",
			exitKey: "p",
			wantContains: []string{
				"COMPLETED",
				"Test task",
				"Week of",
			},
			update: "Should display single todo",
		},
		{
			name: "multiple todos same day",
			todos: []Todo{
				{Text: "Task 1", CreatedAt: now, CompletedAt: &completedNow},
				{Text: "Task 2", CreatedAt: now, CompletedAt: &completedNow},
			},
			title:   "COMPLETED",
			exitKey: "x",
			wantContains: []string{
				"Task 1",
				"Task 2",
				"Week of",
			},
			update: "Should display multiple todos from same day",
		},
		{
			name: "todos with updates",
			todos: []Todo{
				{
					Text:        "Task with notes",
					Updates:     []string{"Note 1", "Note 2"},
					CreatedAt:   now,
					CompletedAt: &completedNow,
				},
			},
			title:   "VIEW",
			exitKey: "p",
			wantContains: []string{
				"Task with notes",
				"Note 1",
				"Note 2",
			},
			update: "Should display updates",
		},
		{
			name: "todos from different days",
			todos: []Todo{
				{Text: "Today task", CreatedAt: now, CompletedAt: &completedNow},
				{Text: "Yesterday task", CreatedAt: yesterday, CompletedAt: &completedYesterday},
			},
			title:   "MULTI DAY",
			exitKey: "p",
			wantContains: []string{
				"Today task",
				"Yesterday task",
				"Week of",
			},
			update: "Should group by day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				completed: tt.todos,
				width:     80,
			}

			result := m.renderPrettifyView(tt.todos, tt.title, tt.exitKey)

			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf("renderPrettifyView() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderWrappedTextWithCursor(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		cursorPos   int
		maxWidth    int
		expectedLen int
		update      string
	}{
		{
			name:        "short text no wrapping",
			text:        "hello",
			cursorPos:   2,
			maxWidth:    10,
			expectedLen: 1,
			update:      "Text shorter than maxWidth should return single line",
		},
		{
			name:        "long text with wrapping",
			text:        "This is a very long text that needs wrapping",
			cursorPos:   5,
			maxWidth:    15,
			expectedLen: 3,
			update:      "Long text should wrap to multiple lines",
		},
		{
			name:        "maxWidth zero",
			text:        "test",
			cursorPos:   2,
			maxWidth:    0,
			expectedLen: 1,
			update:      "MaxWidth <= 0 should return single line",
		},
		{
			name:        "cursor beyond text",
			text:        "hi",
			cursorPos:   10,
			maxWidth:    20,
			expectedLen: 1,
			update:      "Cursor beyond text should be handled",
		},
		{
			name:        "empty text",
			text:        "",
			cursorPos:   0,
			maxWidth:    10,
			expectedLen: 1,
			update:      "Empty text should return cursor only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderWrappedTextWithCursor(tt.text, tt.cursorPos, tt.maxWidth)

			if len(result) != tt.expectedLen {
				t.Errorf("renderWrappedTextWithCursor() returned %d lines, want %d",
					len(result), tt.expectedLen)
			}

			// Verify result is not empty
			for i, line := range result {
				if line == "" {
					t.Errorf("renderWrappedTextWithCursor() line[%d] should not be empty", i)
				}
			}
		})
	}
}
