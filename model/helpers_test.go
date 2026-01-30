package model

import (
	"os"
	"path/filepath"
	"strings"
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

func TestTruncateToDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantHour int
		wantMin  int
		wantSec  int
	}{
		{
			name:     "morning time",
			input:    time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
		},
		{
			name:     "midnight",
			input:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
		},
		{
			name:     "end of day",
			input:    time.Date(2024, 1, 15, 23, 59, 59, 999, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateToDay(tt.input)
			if result.Hour() != tt.wantHour || result.Minute() != tt.wantMin || result.Second() != tt.wantSec {
				t.Errorf("truncateToDay() = %v, want hour=%d min=%d sec=%d",
					result, tt.wantHour, tt.wantMin, tt.wantSec)
			}
			// Verify date is preserved
			if result.Year() != tt.input.Year() || result.Month() != tt.input.Month() || result.Day() != tt.input.Day() {
				t.Errorf("truncateToDay() changed the date: got %v, want %v", result.Format("2006-01-02"), tt.input.Format("2006-01-02"))
			}
		})
	}
}

func TestGetWeekStart(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantDay  int
		wantWeek time.Weekday
	}{
		{
			name:     "Sunday should stay Sunday",
			input:    time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC), // Sunday
			wantDay:  14,
			wantWeek: time.Sunday,
		},
		{
			name:     "Monday should go to previous Sunday",
			input:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday
			wantDay:  14,
			wantWeek: time.Sunday,
		},
		{
			name:     "Saturday should go to previous Sunday",
			input:    time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC), // Saturday
			wantDay:  14,
			wantWeek: time.Sunday,
		},
		{
			name:     "Wednesday in middle of week",
			input:    time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC), // Wednesday
			wantDay:  14,
			wantWeek: time.Sunday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWeekStart(tt.input)
			if result.Weekday() != tt.wantWeek {
				t.Errorf("getWeekStart() weekday = %v, want %v", result.Weekday(), tt.wantWeek)
			}
			if result.Day() != tt.wantDay {
				t.Errorf("getWeekStart() day = %d, want %d", result.Day(), tt.wantDay)
			}
			// Should be truncated to start of day
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
				t.Errorf("getWeekStart() not truncated to day start: %v", result)
			}
		})
	}
}

func TestFormatWeekRange(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  string
	}{
		{
			name:  "same month",
			start: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			want:  "Jan 14 - 20",
		},
		{
			name:  "different months",
			start: time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 2, 3, 0, 0, 0, 0, time.UTC),
			want:  "Jan 28 - Feb 3",
		},
		{
			name:  "year boundary",
			start: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
			want:  "Dec 31 - Jan 6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatWeekRange(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("formatWeekRange() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDayHeader(t *testing.T) {
	now := time.Now()
	today := truncateToDay(now)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	tests := []struct {
		name      string
		input     time.Time
		wantMatch string
	}{
		{
			name:      "today",
			input:     today,
			wantMatch: "Today (",
		},
		{
			name:      "yesterday",
			input:     yesterday,
			wantMatch: "Yesterday (",
		},
		{
			name:      "two days ago",
			input:     twoDaysAgo,
			wantMatch: twoDaysAgo.Format("Monday"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDayHeader(tt.input)
			if !strings.Contains(got, tt.wantMatch) {
				t.Errorf("formatDayHeader() = %q, want to contain %q", got, tt.wantMatch)
			}
		})
	}
}

func TestGroupTodosByWeek(t *testing.T) {
	// Create test times for specific days
	sun := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC) // Sunday
	mon := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Monday
	tue := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC) // Tuesday
	nextWeekSun := time.Date(2024, 1, 21, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		todos       []Todo
		wantWeeks   int
		wantDays    int
		description string
	}{
		{
			name:        "empty todos",
			todos:       []Todo{},
			wantWeeks:   0,
			wantDays:    0,
			description: "Should return empty slice",
		},
		{
			name: "single todo",
			todos: []Todo{
				{Text: "task1", CreatedAt: sun, CompletedAt: &sun},
			},
			wantWeeks:   1,
			wantDays:    1,
			description: "Single todo should create one week, one day",
		},
		{
			name: "multiple todos same day",
			todos: []Todo{
				{Text: "task1", CreatedAt: sun, CompletedAt: &sun},
				{Text: "task2", CreatedAt: sun, CompletedAt: &sun},
			},
			wantWeeks:   1,
			wantDays:    1,
			description: "Same day todos should be in same day group",
		},
		{
			name: "multiple todos same week",
			todos: []Todo{
				{Text: "sunday", CreatedAt: sun, CompletedAt: &sun},
				{Text: "monday", CreatedAt: mon, CompletedAt: &mon},
				{Text: "tuesday", CreatedAt: tue, CompletedAt: &tue},
			},
			wantWeeks:   1,
			wantDays:    3,
			description: "Same week todos should be in same week with multiple days",
		},
		{
			name: "multiple weeks",
			todos: []Todo{
				{Text: "week1", CreatedAt: sun, CompletedAt: &sun},
				{Text: "week2", CreatedAt: nextWeekSun, CompletedAt: &nextWeekSun},
			},
			wantWeeks:   2,
			wantDays:    1, // Each week has 1 day
			description: "Different weeks should create separate week groups",
		},
		{
			name: "todos with nil CompletedAt",
			todos: []Todo{
				{Text: "completed", CreatedAt: sun, CompletedAt: &sun},
				{Text: "not completed", CreatedAt: mon, CompletedAt: nil},
			},
			wantWeeks:   1,
			wantDays:    1,
			description: "Nil CompletedAt should be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weeks := groupTodosByWeek(tt.todos)

			if len(weeks) != tt.wantWeeks {
				t.Errorf("groupTodosByWeek() weeks = %d, want %d", len(weeks), tt.wantWeeks)
			}

			if tt.wantWeeks > 0 {
				totalDays := 0
				for _, week := range weeks {
					totalDays += len(week.Days)
				}
				if tt.wantDays > 0 && totalDays < tt.wantDays {
					t.Errorf("groupTodosByWeek() total days = %d, want at least %d", totalDays, tt.wantDays)
				}
			}
		})
	}
}

func TestGenerateMarkdownFromTodos(t *testing.T) {
	now := time.Now()
	todo1 := now.Add(-2 * time.Hour)
	todo2 := now.Add(-1 * time.Hour)

	tests := []struct {
		name            string
		todos           []Todo
		includeBackups  bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:  "empty todos",
			todos: []Todo{},
			wantContains: []string{
				"# Completed Todos",
				"No completed todos found",
			},
		},
		{
			name: "single todo without description",
			todos: []Todo{
				{Text: "Test Task", CreatedAt: now, CompletedAt: &todo1},
			},
			wantContains: []string{
				"# Completed Todos",
				"**Total completed todos:** 1",
				"Test Task",
			},
		},
		{
			name: "todo with description",
			todos: []Todo{
				{
					Text:        "Task with notes",
					CreatedAt:   now,
					CompletedAt: &todo1,
					Description: []string{"Note 1", "Note 2"},
				},
			},
			wantContains: []string{
				"Task with notes",
				"Note 1",
				"Note 2",
			},
		},
		{
			name: "multiple todos",
			todos: []Todo{
				{Text: "First Task", CreatedAt: now, CompletedAt: &todo1},
				{Text: "Second Task", CreatedAt: now, CompletedAt: &todo2},
			},
			wantContains: []string{
				"**Total completed todos:** 2",
				"First Task",
				"Second Task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMarkdownFromTodos(tt.todos, tt.includeBackups)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("generateMarkdownFromTodos() missing %q", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("generateMarkdownFromTodos() should not contain %q", notWant)
				}
			}
		})
	}
}

func TestExportMarkdownFile(t *testing.T) {
	now := time.Now()
	todo1 := now
	todos := []Todo{
		{Text: "Test Task", CreatedAt: now, CompletedAt: &todo1},
	}

	// Clean up any files created during test
	defer func() {
		// Remove any markdown files created during test
		matches, _ := filepath.Glob("completed_todos_*.md")
		for _, match := range matches {
			_ = os.Remove(match)
		}
	}()

	filename, err := exportMarkdownFile(todos, false)
	if err != nil {
		t.Fatalf("exportMarkdownFile() error = %v", err)
	}

	if filename == "" {
		t.Error("exportMarkdownFile() returned empty filename")
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("exportMarkdownFile() did not create file %q", filename)
	}

	// Verify file contains expected content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test Task") {
		t.Error("Exported markdown does not contain todo text")
	}
	if !strings.Contains(contentStr, "# Completed Todos") {
		t.Error("Exported markdown does not contain header")
	}
}
