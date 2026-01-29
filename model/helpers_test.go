package model

import (
	"os"
	"testing"
	"time"
)

func TestGetCurrentList(t *testing.T) {
	m := Model{
		backlog: []Todo{
			{Text: "backlog1", CreatedAt: time.Now()},
		},
		ready: []Todo{
			{Text: "ready1", CreatedAt: time.Now()},
		},
		displayedCompleted: []Todo{
			{Text: "completed1", CreatedAt: time.Now()},
		},
	}

	tests := []struct {
		name         string
		view         view
		expectedLen  int
		expectedText string
	}{
		{"backlog view", viewBacklog, 1, "backlog1"},
		{"ready view", viewReady, 1, "ready1"},
		{"completed view", viewCompleted, 1, "completed1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.currentView = tt.view
			list := m.getCurrentList()
			if len(list) != tt.expectedLen {
				t.Errorf("getCurrentList() len = %d, want %d", len(list), tt.expectedLen)
			}
			if len(list) > 0 && list[0].Text != tt.expectedText {
				t.Errorf("getCurrentList()[0].Text = %q, want %q", list[0].Text, tt.expectedText)
			}
		})
	}
}

func TestUpdateDisplayedCompleted(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name          string
		completed     []Todo
		expectedLen   int
		expectedFirst string
		expectedOrder []string
	}{
		{
			name:          "empty list",
			completed:     []Todo{},
			expectedLen:   0,
			expectedFirst: "",
			expectedOrder: []string{},
		},
		{
			name: "single todo",
			completed: []Todo{
				{Text: "task1", CreatedAt: now, CompletedAt: &now},
			},
			expectedLen:   1,
			expectedFirst: "task1",
			expectedOrder: []string{"task1"},
		},
		{
			name: "multiple todos sorted by completion time",
			completed: []Todo{
				{Text: "oldest", CreatedAt: now, CompletedAt: &past},
				{Text: "newest", CreatedAt: now, CompletedAt: &future},
				{Text: "middle", CreatedAt: now, CompletedAt: &now},
			},
			expectedLen:   3,
			expectedFirst: "newest",
			expectedOrder: []string{"newest", "middle", "oldest"},
		},
		{
			name: "more than 10 todos",
			completed: func() []Todo {
				todos := make([]Todo, 15)
				for i := 0; i < 15; i++ {
					completedTime := now.Add(time.Duration(i) * time.Minute)
					todos[i] = Todo{
						Text:        string(rune('a' + i)),
						CreatedAt:   now,
						CompletedAt: &completedTime,
					}
				}
				return todos
			}(),
			expectedLen:   10,
			expectedFirst: "o", // 15th item (index 14, char 'o')
			expectedOrder: []string{"o", "n", "m", "l", "k", "j", "i", "h", "g", "f"},
		},
		{
			name: "todos with nil CompletedAt",
			completed: []Todo{
				{Text: "completed", CreatedAt: now, CompletedAt: &now},
				{Text: "notcompleted", CreatedAt: now, CompletedAt: nil},
			},
			expectedLen:   2,
			expectedFirst: "completed",
			expectedOrder: []string{"completed", "notcompleted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				completed: tt.completed,
			}
			m.updateDisplayedCompleted()

			if len(m.displayedCompleted) != tt.expectedLen {
				t.Errorf("updateDisplayedCompleted() len = %d, want %d", len(m.displayedCompleted), tt.expectedLen)
			}

			if tt.expectedLen > 0 {
				if m.displayedCompleted[0].Text != tt.expectedFirst {
					t.Errorf("updateDisplayedCompleted()[0].Text = %q, want %q", m.displayedCompleted[0].Text, tt.expectedFirst)
				}

				// Verify order
				for i, expectedText := range tt.expectedOrder {
					if i < len(m.displayedCompleted) && m.displayedCompleted[i].Text != expectedText {
						t.Errorf("updateDisplayedCompleted()[%d].Text = %q, want %q", i, m.displayedCompleted[i].Text, expectedText)
					}
				}
			}
		})
	}
}

func TestUpdateCompletedTodo(t *testing.T) {
	now := time.Now()
	m := Model{
		completed: []Todo{
			{Text: "task1", CreatedAt: now, CompletedAt: &now},
			{Text: "task2", CreatedAt: now, CompletedAt: &now},
		},
		cursor: 0,
	}
	m.updateDisplayedCompleted()

	// Update the first todo
	m.updateCompletedTodo(func(t *Todo) {
		t.Description = []string{"updated description"}
	})

	if len(m.completed[0].Description) != 1 || m.completed[0].Description[0] != "updated description" {
		t.Errorf("updateCompletedTodo() did not update description, got %v", m.completed[0].Description)
	}
}

func TestSwapTodos(t *testing.T) {
	now := time.Now()
	todos := []Todo{
		{Text: "task1", CreatedAt: now},
		{Text: "task2", CreatedAt: now},
		{Text: "task3", CreatedAt: now},
	}

	// Create a temporary file for testing
	tmpfile := "/tmp/test_swap_todos.txt"
	defer func() {
		_ = os.Remove(tmpfile)
	}()

	// Swap first two items
	swapTodos(todos, 0, 1, tmpfile)

	if todos[0].Text != "task2" {
		t.Errorf("todos[0].Text after swap = %q, want 'task2'", todos[0].Text)
	}
	if todos[1].Text != "task1" {
		t.Errorf("todos[1].Text after swap = %q, want 'task1'", todos[1].Text)
	}

	// Verify file was saved
	loaded := loadTodos(tmpfile)
	if len(loaded) != 3 {
		t.Errorf("loaded todos length = %d, want 3", len(loaded))
	}
	if len(loaded) > 0 && loaded[0].Text != "task2" {
		t.Errorf("loaded[0].Text = %q, want 'task2'", loaded[0].Text)
	}
}

func TestCountCompletedToday(t *testing.T) {
	now := time.Now()
	today := now
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	tests := []struct {
		name          string
		completed     []Todo
		expectedCount int
		description   string
	}{
		{
			name:          "no completed todos",
			completed:     []Todo{},
			expectedCount: 0,
			description:   "Empty list should return 0",
		},
		{
			name: "all completed today",
			completed: []Todo{
				{Text: "task1", CreatedAt: today, CompletedAt: &today},
				{Text: "task2", CreatedAt: today, CompletedAt: &today},
				{Text: "task3", CreatedAt: today, CompletedAt: &today},
			},
			expectedCount: 3,
			description:   "All tasks completed today",
		},
		{
			name: "mixed today and yesterday",
			completed: []Todo{
				{Text: "today1", CreatedAt: today, CompletedAt: &today},
				{Text: "yesterday1", CreatedAt: yesterday, CompletedAt: &yesterday},
				{Text: "today2", CreatedAt: today, CompletedAt: &today},
			},
			expectedCount: 2,
			description:   "Should count only today's completions",
		},
		{
			name: "none completed today",
			completed: []Todo{
				{Text: "yesterday1", CreatedAt: yesterday, CompletedAt: &yesterday},
				{Text: "2days", CreatedAt: twoDaysAgo, CompletedAt: &twoDaysAgo},
			},
			expectedCount: 0,
			description:   "Old completions should not be counted",
		},
		{
			name: "todos with nil CompletedAt",
			completed: []Todo{
				{Text: "completed", CreatedAt: today, CompletedAt: &today},
				{Text: "not completed", CreatedAt: today, CompletedAt: nil},
			},
			expectedCount: 1,
			description:   "Nil CompletedAt should be ignored",
		},
		{
			name: "timezone edge case - just before midnight",
			completed: []Todo{
				{Text: "late", CreatedAt: today, CompletedAt: &today},
			},
			expectedCount: 1,
			description:   "Same day regardless of time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				completed: tt.completed,
			}
			count := m.countCompletedToday()

			if count != tt.expectedCount {
				t.Errorf("countCompletedToday() = %d, want %d", count, tt.expectedCount)
			}
		})
	}
}
