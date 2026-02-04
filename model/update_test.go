package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateCursorNavigation(t *testing.T) {
	m := Model{
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
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("cursor after 'j' = %d, want 1", m.cursor)
	}

	// Move down again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.cursor != 2 {
		t.Errorf("cursor after second 'j' = %d, want 2", m.cursor)
	}

	// Try to move down past end (should stay at 2)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.cursor != 2 {
		t.Errorf("cursor after 'j' at end = %d, want 2", m.cursor)
	}

	// Move up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("cursor after 'k' = %d, want 1", m.cursor)
	}

	// Move up to beginning
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor after second 'k' = %d, want 0", m.cursor)
	}

	// Try to move up past beginning (should stay at 0)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor after 'k' at beginning = %d, want 0", m.cursor)
	}
}

func TestUpdateViewSwitching(t *testing.T) {
	m := Model{
		currentView: viewReady,
		cursor:      5,
	}

	// Switch left to backlog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(Model)
	if m.currentView != viewBacklog {
		t.Errorf("currentView after 'h' from Ready = %v, want %v", m.currentView, viewBacklog)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after view switch = %d, want 0", m.cursor)
	}

	// Switch right to ready
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(Model)
	if m.currentView != viewReady {
		t.Errorf("currentView after 'l' from Backlog = %v, want %v", m.currentView, viewReady)
	}

	// Switch right to completed
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(Model)
	if m.currentView != viewCompleted {
		t.Errorf("currentView after 'l' from Ready = %v, want %v", m.currentView, viewCompleted)
	}

	// Try to switch right from completed (should stay at completed)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(Model)
	if m.currentView != viewCompleted {
		t.Errorf("currentView after 'l' from Completed = %v, want %v", m.currentView, viewCompleted)
	}

	// Switch left to ready
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(Model)
	if m.currentView != viewReady {
		t.Errorf("currentView after 'h' from Completed = %v, want %v", m.currentView, viewReady)
	}
}

func TestUpdateAddingMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog:     []Todo{},
	}

	// Enter adding mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	if !m.adding {
		t.Error("adding mode not activated after 'a'")
	}

	// Type some text
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if m.newTodo != "hi" {
		t.Errorf("newTodo = %q, want 'hi'", m.newTodo)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.adding {
		t.Error("adding mode still active after escape")
	}
	if m.newTodo != "" {
		t.Errorf("newTodo after escape = %q, want empty", m.newTodo)
	}
}

func TestUpdateAddingToTopMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "existing task", CreatedAt: time.Now()},
		},
	}

	// Enter adding to top mode with 'A'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = updated.(Model)
	if !m.adding {
		t.Error("adding mode not activated after 'A'")
	}
	if !m.addingToTop {
		t.Error("addingToTop flag not set after 'A'")
	}

	// Type some text
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(Model)
	if m.newTodo != "new" {
		t.Errorf("newTodo = %q, want 'new'", m.newTodo)
	}

	// Save the todo
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.adding {
		t.Error("adding mode still active after enter")
	}
	if m.addingToTop {
		t.Error("addingToTop flag still set after enter")
	}

	// Check that the new todo was added to the top
	if len(m.backlog) != 2 {
		t.Errorf("backlog length = %d, want 2", len(m.backlog))
	}
	if m.backlog[0].Text != "New" { // Should be capitalized
		t.Errorf("first todo = %q, want 'New'", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "existing task" {
		t.Errorf("second todo = %q, want 'existing task'", m.backlog[1].Text)
	}
	// Cursor should be at the top (position 0)
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}

	// Test with Ready view as well
	m = Model{
		currentView: viewReady,
		ready: []Todo{
			{Text: "existing ready task", CreatedAt: time.Now()},
		},
	}

	// Add to top in ready view
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if len(m.ready) != 2 {
		t.Errorf("ready length = %d, want 2", len(m.ready))
	}
	if m.ready[0].Text != "Top" {
		t.Errorf("first ready todo = %q, want 'Top'", m.ready[0].Text)
	}
	if m.ready[1].Text != "existing ready task" {
		t.Errorf("second ready todo = %q, want 'existing ready task'", m.ready[1].Text)
	}

	// Test cancelling with escape - should reset addingToTop flag
	m = Model{
		currentView: viewBacklog,
		backlog:     []Todo{},
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.addingToTop {
		t.Error("addingToTop flag still set after escape")
	}
}

func TestUpdateEditingUpdatesMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Enter editing mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)
	if !m.editingUpdate {
		t.Error("editing mode not activated after 'u'")
	}

	// Type some text
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)
	if m.newUpdate != "desc" {
		t.Errorf("newUpdate = %q, want 'desc'", m.newUpdate)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.editingUpdate {
		t.Error("editing mode still active after escape")
	}
	if m.newUpdate != "" {
		t.Errorf("newUpdate after escape = %q, want empty", m.newUpdate)
	}
}

func TestUpdateRenamingMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "oldname", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Enter renaming mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)
	if !m.renamingTodo {
		t.Error("renaming mode not activated after 'n'")
	}
	if m.newTodoName != "oldname" {
		t.Errorf("newTodoName = %q, want 'oldname'", m.newTodoName)
	}

	// Type new name
	m.newTodoName = "" // Clear for test
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(Model)
	if m.newTodoName != "new" {
		t.Errorf("newTodoName = %q, want 'new'", m.newTodoName)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.renamingTodo {
		t.Error("renaming mode still active after escape")
	}
	if m.newTodoName != "" {
		t.Errorf("newTodoName after escape = %q, want empty", m.newTodoName)
	}
}

func TestUpdateDeleteConfirmation(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)
	if !m.confirmingDelete {
		t.Error("confirmingDelete not activated after 'd'")
	}

	// Press 'n' to cancel
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)
	if m.confirmingDelete {
		t.Error("confirmingDelete still active after 'n'")
	}
	if len(m.backlog) != 1 {
		t.Errorf("backlog length after cancel = %d, want 1", len(m.backlog))
	}
}

func TestUpdateToggleUpdates(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1"}, CreatedAt: time.Now()},
		},
		cursor:        0,
		showingUpdate: false,
	}

	// Toggle update on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if !m.showingUpdate {
		t.Error("showingUpdate should be true after 'i'")
	}

	// Toggle update off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if m.showingUpdate {
		t.Error("showingUpdate should be false after second 'i'")
	}
}

func TestUpdateToggleAllUpdates(t *testing.T) {
	m := Model{
		currentView:       viewBacklog,
		showingAllUpdates: false,
	}

	// Toggle all updates on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(Model)
	if !m.showingAllUpdates {
		t.Error("showingAllUpdates should be true after 'I'")
	}

	// Toggle all updates off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(Model)
	if m.showingAllUpdates {
		t.Error("showingAllUpdates should be false after second 'I'")
	}
}

func TestInitialModel(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name              string
		backlogTodos      []Todo
		readyTodos        []Todo
		completedTodos    []Todo
		expectedBacklog   int
		expectedReady     int
		expectedCompleted int
		update            string
	}{
		{
			name:              "empty state",
			backlogTodos:      []Todo{},
			readyTodos:        []Todo{},
			completedTodos:    []Todo{},
			expectedBacklog:   0,
			expectedReady:     0,
			expectedCompleted: 0,
			update:            "Should initialize with empty lists",
		},
		{
			name: "with existing todos",
			backlogTodos: []Todo{
				{Text: "Backlog task", CreatedAt: now},
			},
			readyTodos: []Todo{
				{Text: "Ready task", CreatedAt: now},
			},
			completedTodos: []Todo{
				{Text: "Completed task", CreatedAt: now, CompletedAt: &completedTime},
			},
			expectedBacklog:   1,
			expectedReady:     1,
			expectedCompleted: 1,
			update:            "Should load existing todos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean directory
			matches, _ := filepath.Glob("*")
			for _, match := range matches {
				os.Remove(match)
			}

			// Create todo files
			if len(tt.backlogTodos) > 0 {
				saveTodos(backlogFile, tt.backlogTodos)
			}
			if len(tt.readyTodos) > 0 {
				saveTodos(readyFile, tt.readyTodos)
			}
			if len(tt.completedTodos) > 0 {
				saveTodos(completedFile, tt.completedTodos)
			}

			// Call InitialModel
			m := InitialModel()

			// Verify lists
			if len(m.backlog) != tt.expectedBacklog {
				t.Errorf("backlog len = %d, want %d", len(m.backlog), tt.expectedBacklog)
			}
			if len(m.ready) != tt.expectedReady {
				t.Errorf("ready len = %d, want %d", len(m.ready), tt.expectedReady)
			}
			if len(m.displayedCompleted) != tt.expectedCompleted {
				t.Errorf("displayedCompleted len = %d, want %d", len(m.displayedCompleted), tt.expectedCompleted)
			}

			// Verify initial state
			if m.currentView != viewReady {
				t.Errorf("currentView = %v, want %v", m.currentView, viewReady)
			}
			if m.cursor != 0 {
				t.Errorf("cursor = %d, want 0", m.cursor)
			}
		})
	}
}

func TestInit(t *testing.T) {
	m := Model{}
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestUpdateMoveBacklogToReady(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		ready:  []Todo{},
		cursor: 0,
	}

	// Press 'r' to move to ready
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(Model)

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
	m := Model{
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
	m = updated.(Model)

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
	m := Model{
		currentView: viewCompleted,
		backlog:     []Todo{},
		ready:       []Todo{},
		completed: []Todo{
			{Text: "task1", CreatedAt: now, CompletedAt: &now},
		},
		cursor: 0,
	}
	m.updateDisplayedCompleted()

	// Press 'r' to move to ready
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(Model)

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
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'J' to swap with item below
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	m = updated.(Model)

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
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 1,
	}

	// Press 'K' to swap with item above
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m = updated.(Model)

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
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	// Press 'y' to confirm delete
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)

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
	m := Model{
		currentView: viewBacklog,
		backlog:     []Todo{},
		adding:      true,
		newTodo:     "new task",
	}

	// Press Enter to save
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

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

func TestUpdateSaveUpdates(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor:        0,
		editingUpdate: true,
		newUpdate:     "new update",
	}

	// Press Enter to save update
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if len(m.backlog[0].Updates) != 1 || m.backlog[0].Updates[0] != "new update" {
		t.Errorf("backlog[0].Updates = %v, want ['new update']", m.backlog[0].Updates)
	}
	if m.editingUpdate {
		t.Error("editingUpdate should be false after save")
	}
	if m.newUpdate != "" {
		t.Errorf("newUpdate should be empty after save, got %q", m.newUpdate)
	}
}

func TestUpdateSaveRename(t *testing.T) {
	m := Model{
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
	m = updated.(Model)

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

func TestUpdateMoveReadyToBacklog(t *testing.T) {
	m := Model{
		currentView: viewReady,
		backlog: []Todo{
			{Text: "existing backlog", CreatedAt: time.Now()},
		},
		ready: []Todo{
			{Text: "task to move", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 'b' to move from ready to backlog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = updated.(Model)

	// Verify task moved
	if len(m.ready) != 0 {
		t.Errorf("ready length after move = %d, want 0", len(m.ready))
	}
	if len(m.backlog) != 2 {
		t.Errorf("backlog length after move = %d, want 2", len(m.backlog))
	}

	// Verify task appears at TOP of backlog (per CLAUDE.md)
	if m.backlog[0].Text != "task to move" {
		t.Errorf("backlog[0].Text = %q, want 'task to move'", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "existing backlog" {
		t.Errorf("backlog[1].Text = %q, want 'existing backlog'", m.backlog[1].Text)
	}
}

func TestUpdateBackupAndClear(t *testing.T) {
	// Change to temporary directory
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)

	m := Model{
		currentView: viewCompleted,
		completed: []Todo{
			{Text: "task1", CreatedAt: now, CompletedAt: &completedTime},
			{Text: "task2", CreatedAt: now, CompletedAt: &completedTime},
		},
		cursor: 0,
	}
	m.updateDisplayedCompleted()

	// Press 'B' to backup and clear
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'B'}})
	m = updated.(Model)

	// Verify completed list is cleared
	if len(m.completed) != 0 {
		t.Errorf("completed length after backup = %d, want 0", len(m.completed))
	}

	// Verify cursor reset
	if m.cursor != 0 {
		t.Errorf("cursor after backup = %d, want 0", m.cursor)
	}

	// Verify success message shown
	if !contains(m.message, "Backed up") {
		t.Errorf("message should contain 'Backed up', got %q", m.message)
	}

	// Verify backup file exists
	dateStr := now.Format("2006-01-02")
	expectedBackupPrefix := "todo_completed_backup_" + dateStr
	files, _ := os.ReadDir(".")
	backupFound := false
	for _, file := range files {
		if len(file.Name()) >= len(expectedBackupPrefix) && file.Name()[:len(expectedBackupPrefix)] == expectedBackupPrefix {
			backupFound = true
			// Clean up
			os.Remove(file.Name())
		}
	}
	if !backupFound {
		t.Error("Backup file was not created")
	}
}

func TestUpdateToggleHelp(t *testing.T) {
	m := Model{
		showingCommands: false,
	}

	// Press '?' to toggle help on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)

	if !m.showingCommands {
		t.Error("showingCommands should be true after '?'")
	}

	// Press '?' again to toggle off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)

	if m.showingCommands {
		t.Error("showingCommands should be false after second '?'")
	}
}

func TestUpdateEnterUpdatesNavigation(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1", "desc2"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: false,
	}

	// Press Enter to enter update navigation mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.navigatingUpdates {
		t.Error("navigatingUpdates should be true after Enter")
	}
	if m.updateCursor != 0 {
		t.Errorf("updateCursor should be 0, got %d", m.updateCursor)
	}
	if !m.showingUpdate {
		t.Error("showingUpdate should be true after Enter")
	}
}

func TestUpdateEnterUpdatesNavigationNoUpdates(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task without desc", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press Enter when todo has no updates
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.navigatingUpdates {
		t.Error("navigatingUpdates should be false for todo with no updates")
	}
	if !contains(m.message, "No updates") {
		t.Errorf("message should indicate no updates, got %q", m.message)
	}
}

func TestUpdatesNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1", "desc2", "desc3"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: true,
		updateCursor:      0,
		showingUpdate:     true,
	}

	// Navigate down with 'j'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.updateCursor != 1 {
		t.Errorf("updateCursor after 'j' = %d, want 1", m.updateCursor)
	}

	// Navigate down again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.updateCursor != 2 {
		t.Errorf("updateCursor after second 'j' = %d, want 2", m.updateCursor)
	}

	// Try to navigate past end (should stay at 2)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.updateCursor != 2 {
		t.Errorf("updateCursor should not go past end, got %d", m.updateCursor)
	}

	// Navigate up with 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.updateCursor != 1 {
		t.Errorf("updateCursor after 'k' = %d, want 1", m.updateCursor)
	}

	// Navigate to top
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.updateCursor != 0 {
		t.Errorf("updateCursor after second 'k' = %d, want 0", m.updateCursor)
	}

	// Try to navigate past beginning (should stay at 0)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.updateCursor != 0 {
		t.Errorf("updateCursor should not go below 0, got %d", m.updateCursor)
	}

	// Exit with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.navigatingUpdates {
		t.Error("navigatingUpdates should be false after Esc")
	}
	if m.updateCursor != 0 {
		t.Errorf("updateCursor should reset to 0, got %d", m.updateCursor)
	}
}

func TestUpdatesNavigationCreateNewUpdate(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1", "desc2"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: true,
		updateCursor:      1,
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)

	if m.navigatingUpdates {
		t.Error("navigatingUpdates should be false after 'u'")
	}
	if !m.editingUpdate {
		t.Error("editingUpdate should be true after 'u'")
	}
	if m.updateCursor != 0 {
		t.Errorf("updateCursor should reset to 0, got %d", m.updateCursor)
	}
	if m.newUpdate != "" {
		t.Errorf("newUpdate should be empty, got %q", m.newUpdate)
	}
}

func TestUpdatesNavigationExitOnTodoChange(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1"}, CreatedAt: time.Now()},
			{Text: "task2", Updates: []string{"desc2"}, CreatedAt: time.Now()},
		},
		cursor:            1,
		navigatingUpdates: false,
		updateCursor:      0,
	}

	// When NOT in navigation mode, moving between todos with 'j'/'k' should ensure navigation stays off
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)

	if m.cursor != 0 {
		t.Errorf("cursor should be 0 after 'k', got %d", m.cursor)
	}
	if m.navigatingUpdates {
		t.Error("navigatingUpdates should be false after moving between todos")
	}

	// Move down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after 'j', got %d", m.cursor)
	}
	if m.navigatingUpdates {
		t.Error("navigatingUpdates should remain false")
	}
}

func TestUpdatesDeletionInNavigation(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1", "desc2", "desc3"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: true,
		updateCursor:      1,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	if !m.confirmingDeleteUpdate {
		t.Error("confirmingDeleteUpdate should be true after 'd'")
	}

	// Press 'y' to confirm deletion
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)

	if m.confirmingDeleteUpdate {
		t.Error("confirmingDeleteUpdate should be false after confirmation")
	}
	if len(m.backlog[0].Updates) != 2 {
		t.Errorf("Updates count after deletion = %d, want 2", len(m.backlog[0].Updates))
	}
	// Verify correct update was deleted (desc2 at index 1)
	if m.backlog[0].Updates[0] != "desc1" || m.backlog[0].Updates[1] != "desc3" {
		t.Errorf("Wrong update deleted, got %v", m.backlog[0].Updates)
	}
}

func TestUpdatesDeletionCancelled(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"desc1", "desc2"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingUpdates:      true,
		updateCursor:           0,
		confirmingDeleteUpdate: true,
	}

	// Press 'n' to cancel deletion
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)

	if m.confirmingDeleteUpdate {
		t.Error("confirmingDeleteUpdate should be false after cancellation")
	}
	if len(m.backlog[0].Updates) != 2 {
		t.Errorf("Updates count should be unchanged, got %d", len(m.backlog[0].Updates))
	}
}

func TestUpdatesDeletionExitNavigationWhenEmpty(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"only desc"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingUpdates:      true,
		updateCursor:           0,
		confirmingDeleteUpdate: true,
	}

	// Delete the only update
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)

	// Should exit navigation mode when no updates remain
	if m.navigatingUpdates {
		t.Error("navigatingUpdates should be false when all updates deleted")
	}
	if len(m.backlog[0].Updates) != 0 {
		t.Errorf("Updates count should be 0, got %d", len(m.backlog[0].Updates))
	}
}

func TestEditUpdatesInNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"old desc 1", "old desc 2"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: true,
		updateCursor:      1,
	}

	// Press 'n' to edit update
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)

	if !m.editingUpdate {
		t.Error("editingUpdate should be true after 'n' in navigation mode")
	}
	// Should load existing update text
	if m.newUpdate != "old desc 2" {
		t.Errorf("newUpdate should be 'old desc 2', got %q", m.newUpdate)
	}

	// Modify the update
	m.newUpdate = "updated desc"

	// Save with Enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.editingUpdate {
		t.Error("editingUpdate should be false after save")
	}
	// Should update existing update, not append
	if len(m.backlog[0].Updates) != 2 {
		t.Errorf("Updates count should remain 2, got %d", len(m.backlog[0].Updates))
	}
	if m.backlog[0].Updates[1] != "updated desc" {
		t.Errorf("Updates[1] should be 'updated desc', got %q", m.backlog[0].Updates[1])
	}
	if m.backlog[0].Updates[0] != "old desc 1" {
		t.Errorf("Updates[0] should be unchanged, got %q", m.backlog[0].Updates[0])
	}
}

func TestEditUpdatesNotInNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Updates: []string{"existing"}, CreatedAt: time.Now()},
		},
		cursor:            0,
		navigatingUpdates: false,
	}

	// Press 'u' when NOT in navigation mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)

	if !m.editingUpdate {
		t.Error("editingUpdate should be true")
	}
	// Should start with empty update (new update)
	if m.newUpdate != "" {
		t.Errorf("newUpdate should be empty for new update, got %q", m.newUpdate)
	}

	// Add new update
	m.newUpdate = "new desc"

	// Save with Enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	// Should append new update
	if len(m.backlog[0].Updates) != 2 {
		t.Errorf("Updates count should be 2, got %d", len(m.backlog[0].Updates))
	}
	if m.backlog[0].Updates[0] != "new desc" {
		t.Errorf("Updates[0] should be 'new desc', got %q", m.backlog[0].Updates[0])
	}
	if m.backlog[0].Updates[1] != "existing" {
		t.Errorf("Updates[1] should be 'existing', got %q", m.backlog[0].Updates[1])
	}
}

func TestUpdateMoveToTopBacklog(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
			{Text: "task3", CreatedAt: time.Now()},
		},
		cursor: 2,
	}

	// Press 't' to move task3 to top
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)

	if m.backlog[0].Text != "task3" {
		t.Errorf("backlog[0].Text after move to top = %q, want 'task3'", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "task1" {
		t.Errorf("backlog[1].Text after move to top = %q, want 'task1'", m.backlog[1].Text)
	}
	if m.backlog[2].Text != "task2" {
		t.Errorf("backlog[2].Text after move to top = %q, want 'task2'", m.backlog[2].Text)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after move to top = %d, want 0", m.cursor)
	}
	if m.message != "Todo moved to top" {
		t.Errorf("message after move to top = %q, want 'Todo moved to top'", m.message)
	}
}

func TestUpdateMoveToTopReady(t *testing.T) {
	m := Model{
		currentView: viewReady,
		ready: []Todo{
			{Text: "ready1", CreatedAt: time.Now()},
			{Text: "ready2", CreatedAt: time.Now()},
			{Text: "ready3", CreatedAt: time.Now()},
		},
		cursor: 1,
	}

	// Press 't' to move ready2 to top
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)

	if m.ready[0].Text != "ready2" {
		t.Errorf("ready[0].Text after move to top = %q, want 'ready2'", m.ready[0].Text)
	}
	if m.ready[1].Text != "ready1" {
		t.Errorf("ready[1].Text after move to top = %q, want 'ready1'", m.ready[1].Text)
	}
	if m.ready[2].Text != "ready3" {
		t.Errorf("ready[2].Text after move to top = %q, want 'ready3'", m.ready[2].Text)
	}
	if m.cursor != 0 {
		t.Errorf("cursor after move to top = %d, want 0", m.cursor)
	}
}

func TestUpdateMoveToTopAlreadyAtTop(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
			{Text: "task2", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 't' when already at top (should not change anything)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)

	// Order should remain unchanged
	if m.backlog[0].Text != "task1" {
		t.Errorf("backlog[0].Text should remain 'task1', got %q", m.backlog[0].Text)
	}
	if m.backlog[1].Text != "task2" {
		t.Errorf("backlog[1].Text should remain 'task2', got %q", m.backlog[1].Text)
	}
	if m.cursor != 0 {
		t.Errorf("cursor should remain 0, got %d", m.cursor)
	}
}

func TestUpdateMoveToTopInCompletedView(t *testing.T) {
	now := time.Now()
	m := Model{
		currentView: viewCompleted,
		completed: []Todo{
			{Text: "completed1", CreatedAt: now, CompletedAt: &now},
			{Text: "completed2", CreatedAt: now, CompletedAt: &now},
		},
		cursor: 1,
	}
	m.updateDisplayedCompleted()

	// Press 't' in completed view (should not do anything)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)

	// Nothing should change
	if m.cursor != 1 {
		t.Errorf("cursor should remain 1 in completed view, got %d", m.cursor)
	}
}

func TestUpdateMoveToTopSingleItem(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "only task", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press 't' with only one item (should not change anything)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)

	if len(m.backlog) != 1 {
		t.Errorf("backlog length should remain 1, got %d", len(m.backlog))
	}
	if m.backlog[0].Text != "only task" {
		t.Errorf("backlog[0].Text should remain 'only task', got %q", m.backlog[0].Text)
	}
}
