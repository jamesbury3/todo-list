package model

import (
	"os"
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

func TestUpdateEditingDescriptionMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Enter editing mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	if !m.editingDescription {
		t.Error("editing mode not activated after 'e'")
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
	if m.newDescription != "desc" {
		t.Errorf("newDescription = %q, want 'desc'", m.newDescription)
	}

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.editingDescription {
		t.Error("editing mode still active after escape")
	}
	if m.newDescription != "" {
		t.Errorf("newDescription after escape = %q, want empty", m.newDescription)
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

func TestUpdateToggleDescription(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1"}, CreatedAt: time.Now()},
		},
		cursor:             0,
		showingDescription: false,
	}

	// Toggle description on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if !m.showingDescription {
		t.Error("showingDescription should be true after 'i'")
	}

	// Toggle description off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if m.showingDescription {
		t.Error("showingDescription should be false after second 'i'")
	}
}

func TestUpdateToggleAllDescriptions(t *testing.T) {
	m := Model{
		currentView:            viewBacklog,
		showingAllDescriptions: false,
	}

	// Toggle all descriptions on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(Model)
	if !m.showingAllDescriptions {
		t.Error("showingAllDescriptions should be true after 'I'")
	}

	// Toggle all descriptions off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
	m = updated.(Model)
	if m.showingAllDescriptions {
		t.Error("showingAllDescriptions should be false after second 'I'")
	}
}

func TestInitialModel(t *testing.T) {
	// Test file constants - separate from production files
	const (
		testBacklogFile   = "test_todo_backlog.txt"
		testReadyFile     = "test_todo_ready.txt"
		testCompletedFile = "test_todo_completed.txt"
	)

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
	m := Model{
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

	// Press 'u' to undo completion
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
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

func TestUpdateSaveDescription(t *testing.T) {
	m := Model{
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
	m = updated.(Model)

	if len(m.backlog[0].Description) != 1 || m.backlog[0].Description[0] != "new description" {
		t.Errorf("backlog[0].Description = %v, want ['new description']", m.backlog[0].Description)
	}
	if m.editingDescription {
		t.Error("editingDescription should be false after save")
	}
	if m.newDescription != "" {
		t.Errorf("newDescription should be empty after save, got %q", m.newDescription)
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

func TestUpdateEnterDescriptionNavigation(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1", "desc2"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: false,
	}

	// Press Enter to enter description navigation mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.navigatingDescriptions {
		t.Error("navigatingDescriptions should be true after Enter")
	}
	if m.descriptionCursor != 0 {
		t.Errorf("descriptionCursor should be 0, got %d", m.descriptionCursor)
	}
	if !m.showingDescription {
		t.Error("showingDescription should be true after Enter")
	}
}

func TestUpdateEnterDescriptionNavigationNoDescriptions(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task without desc", CreatedAt: time.Now()},
		},
		cursor: 0,
	}

	// Press Enter when todo has no descriptions
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.navigatingDescriptions {
		t.Error("navigatingDescriptions should be false for todo with no descriptions")
	}
	if !contains(m.message, "No descriptions") {
		t.Errorf("message should indicate no descriptions, got %q", m.message)
	}
}

func TestDescriptionNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1", "desc2", "desc3"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: true,
		descriptionCursor:      0,
		showingDescription:     true,
	}

	// Navigate down with 'j'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.descriptionCursor != 1 {
		t.Errorf("descriptionCursor after 'j' = %d, want 1", m.descriptionCursor)
	}

	// Navigate down again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.descriptionCursor != 2 {
		t.Errorf("descriptionCursor after second 'j' = %d, want 2", m.descriptionCursor)
	}

	// Try to navigate past end (should stay at 2)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.descriptionCursor != 2 {
		t.Errorf("descriptionCursor should not go past end, got %d", m.descriptionCursor)
	}

	// Navigate up with 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.descriptionCursor != 1 {
		t.Errorf("descriptionCursor after 'k' = %d, want 1", m.descriptionCursor)
	}

	// Navigate to top
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.descriptionCursor != 0 {
		t.Errorf("descriptionCursor after second 'k' = %d, want 0", m.descriptionCursor)
	}

	// Try to navigate past beginning (should stay at 0)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.descriptionCursor != 0 {
		t.Errorf("descriptionCursor should not go below 0, got %d", m.descriptionCursor)
	}

	// Exit with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.navigatingDescriptions {
		t.Error("navigatingDescriptions should be false after Esc")
	}
	if m.descriptionCursor != 0 {
		t.Errorf("descriptionCursor should reset to 0, got %d", m.descriptionCursor)
	}
}

func TestDescriptionNavigationExitOnTodoChange(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1"}, CreatedAt: time.Now()},
			{Text: "task2", Description: []string{"desc2"}, CreatedAt: time.Now()},
		},
		cursor:                 1,
		navigatingDescriptions: false,
		descriptionCursor:      0,
	}

	// When NOT in navigation mode, moving between todos with 'j'/'k' should ensure navigation stays off
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)

	if m.cursor != 0 {
		t.Errorf("cursor should be 0 after 'k', got %d", m.cursor)
	}
	if m.navigatingDescriptions {
		t.Error("navigatingDescriptions should be false after moving between todos")
	}

	// Move down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after 'j', got %d", m.cursor)
	}
	if m.navigatingDescriptions {
		t.Error("navigatingDescriptions should remain false")
	}
}

func TestDescriptionDeletionInNavigation(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1", "desc2", "desc3"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: true,
		descriptionCursor:      1,
	}

	// Press 'd' to enter delete confirmation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	if !m.confirmingDeleteDesc {
		t.Error("confirmingDeleteDesc should be true after 'd'")
	}

	// Press 'y' to confirm deletion
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)

	if m.confirmingDeleteDesc {
		t.Error("confirmingDeleteDesc should be false after confirmation")
	}
	if len(m.backlog[0].Description) != 2 {
		t.Errorf("Description count after deletion = %d, want 2", len(m.backlog[0].Description))
	}
	// Verify correct description was deleted (desc2 at index 1)
	if m.backlog[0].Description[0] != "desc1" || m.backlog[0].Description[1] != "desc3" {
		t.Errorf("Wrong description deleted, got %v", m.backlog[0].Description)
	}
}

func TestDescriptionDeletionCancelled(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"desc1", "desc2"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: true,
		descriptionCursor:      0,
		confirmingDeleteDesc:   true,
	}

	// Press 'n' to cancel deletion
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(Model)

	if m.confirmingDeleteDesc {
		t.Error("confirmingDeleteDesc should be false after cancellation")
	}
	if len(m.backlog[0].Description) != 2 {
		t.Errorf("Description count should be unchanged, got %d", len(m.backlog[0].Description))
	}
}

func TestDescriptionDeletionExitNavigationWhenEmpty(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"only desc"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: true,
		descriptionCursor:      0,
		confirmingDeleteDesc:   true,
	}

	// Delete the only description
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)

	// Should exit navigation mode when no descriptions remain
	if m.navigatingDescriptions {
		t.Error("navigatingDescriptions should be false when all descriptions deleted")
	}
	if len(m.backlog[0].Description) != 0 {
		t.Errorf("Description count should be 0, got %d", len(m.backlog[0].Description))
	}
}

func TestEditDescriptionInNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"old desc 1", "old desc 2"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: true,
		descriptionCursor:      1,
	}

	// Press 'e' to edit description
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)

	if !m.editingDescription {
		t.Error("editingDescription should be true after 'e' in navigation mode")
	}
	// Should load existing description text
	if m.newDescription != "old desc 2" {
		t.Errorf("newDescription should be 'old desc 2', got %q", m.newDescription)
	}

	// Modify the description
	m.newDescription = "updated desc"

	// Save with Enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.editingDescription {
		t.Error("editingDescription should be false after save")
	}
	// Should update existing description, not append
	if len(m.backlog[0].Description) != 2 {
		t.Errorf("Description count should remain 2, got %d", len(m.backlog[0].Description))
	}
	if m.backlog[0].Description[1] != "updated desc" {
		t.Errorf("Description[1] should be 'updated desc', got %q", m.backlog[0].Description[1])
	}
	if m.backlog[0].Description[0] != "old desc 1" {
		t.Errorf("Description[0] should be unchanged, got %q", m.backlog[0].Description[0])
	}
}

func TestEditDescriptionNotInNavigationMode(t *testing.T) {
	m := Model{
		currentView: viewBacklog,
		backlog: []Todo{
			{Text: "task1", Description: []string{"existing"}, CreatedAt: time.Now()},
		},
		cursor:                 0,
		navigatingDescriptions: false,
	}

	// Press 'e' when NOT in navigation mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)

	if !m.editingDescription {
		t.Error("editingDescription should be true")
	}
	// Should start with empty description (new description)
	if m.newDescription != "" {
		t.Errorf("newDescription should be empty for new description, got %q", m.newDescription)
	}

	// Add new description
	m.newDescription = "new desc"

	// Save with Enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	// Should append new description
	if len(m.backlog[0].Description) != 2 {
		t.Errorf("Description count should be 2, got %d", len(m.backlog[0].Description))
	}
	if m.backlog[0].Description[0] != "new desc" {
		t.Errorf("Description[0] should be 'new desc', got %q", m.backlog[0].Description[0])
	}
	if m.backlog[0].Description[1] != "existing" {
		t.Errorf("Description[1] should be 'existing', got %q", m.backlog[0].Description[1])
	}
}
