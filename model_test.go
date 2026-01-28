package main

import (
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Test file constants - separate from production files
const (
	testBacklogFile   = "test_todo_backlog.txt"
	testReadyFile     = "test_todo_ready.txt"
	testCompletedFile = "test_todo_completed.txt"
)

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

func TestGetCurrentList(t *testing.T) {
	m := model{
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
			m := model{
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
	m := model{
		completed: []Todo{
			{Text: "task1", CreatedAt: now, CompletedAt: &now},
			{Text: "task2", CreatedAt: now, CompletedAt: &now},
		},
		cursor: 0,
	}
	m.updateDisplayedCompleted()

	// Update the first todo
	m.updateCompletedTodo(func(t *Todo) {
		t.Description = "updated description"
	})

	if m.completed[0].Description != "updated description" {
		t.Errorf("updateCompletedTodo() did not update description, got %q", m.completed[0].Description)
	}
}

func TestUpdateCursorNavigation(t *testing.T) {
	m := model{
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
			{Text: "task3", CreatedAt: time.Now()},
		},
		currentView: viewBacklog,
		cursor:      0,
	}

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.cursor != 1 {
		t.Errorf("cursor after 'j' = %d, want 1", m.cursor)
	}

	// Move down again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.cursor != 2 {
		t.Errorf("cursor after second 'j' = %d, want 2", m.cursor)
	}

	// Try to move down past end (should stay at 2)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.cursor != 2 {
		t.Errorf("cursor after 'j' at end = %d, want 2", m.cursor)
	}

	// Move up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(model)
	if m.cursor != 1 {
		t.Errorf("cursor after 'k' = %d, want 1", m.cursor)
	}

	// Move up to beginning
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(model)
	if m.cursor != 0 {
		t.Errorf("cursor after second 'k' = %d, want 0", m.cursor)
	}

	// Try to move up past beginning (should stay at 0)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(model)
	if m.cursor != 0 {
		t.Errorf("cursor after 'k' at beginning = %d, want 0", m.cursor)
	}
}

func TestUpdateViewSwitching(t *testing.T) {
	m := model{
		currentView: viewReady,
		cursor:      5,
	}

	// Switch left to backlog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(model)
	if m.currentView != viewBacklog {
		t.Errorf("currentView after 'h' from Ready = %v, want %v", m.currentView, viewBacklog)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after view switch = %d, want 0", m.cursor)
	}

	// Switch right to ready
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if m.currentView != viewReady {
		t.Errorf("currentView after 'l' from Backlog = %v, want %v", m.currentView, viewReady)
	}

	// Switch right to completed
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if m.currentView != viewCompleted {
		t.Errorf("currentView after 'l' from Ready = %v, want %v", m.currentView, viewCompleted)
	}

	// Try to switch right from completed (should stay at completed)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if m.currentView != viewCompleted {
		t.Errorf("currentView after 'l' from Completed = %v, want %v", m.currentView, viewCompleted)
	}

	// Switch left to ready
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(model)
	if m.currentView != viewReady {
		t.Errorf("currentView after 'h' from Completed = %v, want %v", m.currentView, viewReady)
	}
}

func TestUpdateAddingMode(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog:     []Todo{},
	}

	// Enter adding mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(model)
	if !m.adding {
		t.Error("adding mode not activated after 'a'")
	}

	// Type some text
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(model)
	if m.newTodo != "hi" {
		t.Errorf("newTodo = %q, want 'hi'", m.newTodo)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.adding {
		t.Error("adding mode still active after escape")
	}
	if m.newTodo != "" {
		t.Errorf("newTodo after escape = %q, want empty", m.newTodo)
	}
}

func TestUpdateEditingDescriptionMode(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Enter editing mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(model)
	if !m.editingDescription {
		t.Error("editing mode not activated after 'e'")
	}

	// Type some text
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if m.newDescription != "desc" {
		t.Errorf("newDescription = %q, want 'desc'", m.newDescription)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.editingDescription {
		t.Error("editing mode still active after escape")
	}
	if m.newDescription != "" {
		t.Errorf("newDescription after escape = %q, want empty", m.newDescription)
	}
}

func TestUpdateRenamingMode(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "oldname", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Enter renaming mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	if !m.renamingTodo {
		t.Error("renaming mode not activated after 'n'")
	}
	if m.newTodoName != "oldname" {
		t.Errorf("newTodoName = %q, want 'oldname'", m.newTodoName)
	}

	// Type new name
	m.newTodoName = "" // Clear for test
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	if m.newTodoName != "new" {
		t.Errorf("newTodoName = %q, want 'new'", m.newTodoName)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.renamingTodo {
		t.Error("renaming mode still active after escape")
	}
	if m.newTodoName != "" {
		t.Errorf("newTodoName after escape = %q, want empty", m.newTodoName)
	}
}

func TestUpdateDeleteConfirmation(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(model)
	if !m.confirmingDelete {
		t.Error("confirmingDelete not activated after 'd'")
	}

	// Press 'n' to cancel
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	if m.confirmingDelete {
		t.Error("confirmingDelete still active after 'n'")
	}
	if len(m.backlog) != 1 {
		t.Errorf("backlog length after cancel = %d, want 1", len(m.backlog))
	}
}

func TestUpdateToggleDescription(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: "desc1", CreatedAt: time.Now()},
		},
		cursor:             0,
		showingDescription: false,
	}

	// Toggle description on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(model)
	if !m.showingDescription {
		t.Error("showingDescription should be true after 'i'")
	}

	// Toggle description off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(model)
	if m.showingDescription {
		t.Error("showingDescription should be false after second 'i'")
	}
}

func TestUpdateToggleAllDescriptions(t *testing.T) {
	m := model{
		currentView:            viewBacklog,
		showingAllDescriptions: false,
	}

	// Toggle all descriptions on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(model)
	if !m.showingAllDescriptions {
		t.Error("showingAllDescriptions should be true after 'I'")
	}

	// Toggle all descriptions off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(model)
	if m.showingAllDescriptions {
		t.Error("showingAllDescriptions should be false after second 'I'")
	}
}

func TestViewRendering(t *testing.T) {
	now := time.Now()
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: now},
			{Text: "task2", Description: "has description", CreatedAt: now},
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
	if !contains(view, "Commands:") {
		t.Error("View should contain commands help")
	}
}

func TestViewRenderingEmptyList(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog:     []Todo{},
	}

	view := m.View()

	if !contains(view, "No todos") {
		t.Error("View should contain 'No todos' for empty list")
	}
}

func TestViewRenderingAddingMode(t *testing.T) {
	m := model{
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
	m := model{
		currentView:      viewBacklog,
		confirmingDelete: true,
	}

	view := m.View()

	if !contains(view, "Are you sure") {
		t.Error("View should contain delete confirmation prompt")
	}
}

func TestInit(t *testing.T) {
	m := model{}
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return tea.ClearScreen command")
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

func TestInitialModel(t *testing.T) {
	// Clean up test files at the end
	defer func() {
		_ = os.Remove(testBacklogFile)
		_ = os.Remove(testReadyFile)
		_ = os.Remove(testCompletedFile)
	}()

	// Create test data
	now := time.Now()
	backlogTodos := []Todo{{Text: "backlog1", CreatedAt: now}}
	readyTodos := []Todo{{Text: "ready1", CreatedAt: now}}
	completedTodos := []Todo{{Text: "completed1", CreatedAt: now, CompletedAt: &now}}

	_ = saveTodos(testBacklogFile, backlogTodos)
	_ = saveTodos(testReadyFile, readyTodos)
	_ = saveTodos(testCompletedFile, completedTodos)

	// Manually construct model with test data (instead of using initialModel() which uses production files)
	m := model{
		backlog:     loadTodos(testBacklogFile),
		ready:       loadTodos(testReadyFile),
		completed:   loadTodos(testCompletedFile),
		cursor:      0,
		currentView: viewReady,
	}
	m.updateDisplayedCompleted()

	// Verify data loaded
	if len(m.backlog) != 1 {
		t.Errorf("backlog length = %d, want 1", len(m.backlog))
	}
	if len(m.ready) != 1 {
		t.Errorf("ready length = %d, want 1", len(m.ready))
	}
	if len(m.completed) != 1 {
		t.Errorf("completed length = %d, want 1", len(m.completed))
	}
	if m.currentView != viewReady {
		t.Errorf("currentView = %v, want %v", m.currentView, viewReady)
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestUpdateMoveBacklogToReady(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		ready:  []Todo{},
		cursor: 0,
	}

	// Press 'r' to move to ready
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(model)

	if len(m.backlog) != 0 {
		t.Errorf("backlog length after move = %d, want 0", len(m.backlog))
	}
	if len(m.ready) != 1 {
		t.Errorf("ready length after move = %d, want 1", len(m.ready))
	}
	if m.ready[0].Text != "task1" {
		t.Errorf("ready[0].Text = %q, want 'task1'", m.ready[0].Text)
	}
}

func TestUpdateMarkComplete(t *testing.T) {
	m := model{
		currentView: viewReady,
		backlog:     []Todo{},
		ready: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		completed: []Todo{},
		cursor:    0,
	}

	// Press 'x' to mark as complete
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(model)

	if len(m.ready) != 0 {
		t.Errorf("ready length after complete = %d, want 0", len(m.ready))
	}
	if len(m.completed) != 1 {
		t.Errorf("completed length after complete = %d, want 1", len(m.completed))
	}
	if m.completed[0].Text != "task1" {
		t.Errorf("completed[0].Text = %q, want 'task1'", m.completed[0].Text)
	}
	if m.completed[0].CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestUpdateUndoComplete(t *testing.T) {
	now := time.Now()
	m := model{
		currentView: viewCompleted,
		backlog:     []Todo{},
		ready:       []Todo{},
		completed: []Todo{
			{Text: "task1", CreatedAt: now, CompletedAt: &now},
		},
		cursor: 0,
	}
	m.updateDisplayedCompleted()

	// Press 'u' to undo completion
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)

	if len(m.completed) != 0 {
		t.Errorf("completed length after undo = %d, want 0", len(m.completed))
	}
	if len(m.ready) != 1 {
		t.Errorf("ready length after undo = %d, want 1", len(m.ready))
	}
	if m.ready[0].Text != "task1" {
		t.Errorf("ready[0].Text = %q, want 'task1'", m.ready[0].Text)
	}
	if m.ready[0].CompletedAt != nil {
		t.Error("CompletedAt should be cleared")
	}
}

func TestUpdateReorderDown(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'J' to swap with item below
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	m = updated.(model)

	if m.backlog[0].Text != "task2" {
		t.Errorf("backlog[0].Text after swap = %q, want 'task2'", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "task1" {
		t.Errorf("backlog[1].Text after swap = %q, want 'task1'", m.backlog[1].Text)
	}
	if m.cursor != 1 {
		t.Errorf("cursor after swap down = %d, want 1", m.cursor)
	}
}

func TestUpdateReorderUp(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 1,
	}

	// Press 'K' to swap with item above
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m = updated.(model)

	if m.backlog[0].Text != "task2" {
		t.Errorf("backlog[0].Text after swap = %q, want 'task2'", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "task1" {
		t.Errorf("backlog[1].Text after swap = %q, want 'task1'", m.backlog[1].Text)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after swap up = %d, want 0", m.cursor)
	}
}

func TestUpdateConfirmDelete(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(model)

	// Press 'y' to confirm delete
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(model)

	if len(m.backlog) != 1 {
		t.Errorf("backlog length after delete = %d, want 1", len(m.backlog))
	}
	if m.backlog[0].Text != "task2" {
		t.Errorf("backlog[0].Text after delete = %q, want 'task2'", m.backlog[0].Text)
	}
	if m.confirmingDelete {
		t.Error("confirmingDelete should be false after deletion")
	}
}

func TestUpdateSaveNewTodo(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog:     []Todo{},
		adding:      true,
		newTodo:     "new task",
	}

	// Press Enter to save
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)

	if len(m.backlog) != 1 {
		t.Errorf("backlog length after save = %d, want 1", len(m.backlog))
	}
	if m.backlog[0].Text != "New task" {
		t.Errorf("backlog[0].Text = %q, want 'new task'", m.backlog[0].Text)
	}
	if m.adding {
		t.Error("adding should be false after save")
	}
	if m.newTodo != "" {
		t.Errorf("newTodo should be empty after save, got %q", m.newTodo)
	}
}

func TestUpdateSaveDescription(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor:             0,
		editingDescription: true,
		newDescription:     "new description",
	}

	// Press Enter to save description
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)

	if m.backlog[0].Description != "new description" {
		t.Errorf("backlog[0].Description = %q, want 'new description'", m.backlog[0].Description)
	}
	if m.editingDescription {
		t.Error("editingDescription should be false after save")
	}
	if m.newDescription != "" {
		t.Errorf("newDescription should be empty after save, got %q", m.newDescription)
	}
}

func TestUpdateSaveRename(t *testing.T) {
	m := model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "old name", CreatedAt: time.Now()},
		},
		cursor:       0,
		renamingTodo: true,
		newTodoName:  "new name",
	}

	// Press Enter to save rename
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)

	if m.backlog[0].Text != "New name" {
		t.Errorf("backlog[0].Text = %q, want 'new name'", m.backlog[0].Text)
	}
	if m.renamingTodo {
		t.Error("renamingTodo should be false after save")
	}
	if m.newTodoName != "" {
		t.Errorf("newTodoName should be empty after save, got %q", m.newTodoName)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
