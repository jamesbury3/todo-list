package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadTodos(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectedCount int
		expectedText  []string
		update        string
	}{
		{
			name:          "empty file",
			fileContent:   "",
			expectedCount: 0,
			expectedText:  []string{},
			update:        "Empty file should return empty slice",
		},
		{
			name:          "single valid JSON",
			fileContent:   `{"text":"Task 1","created_at":"2024-01-01T10:00:00Z"}`,
			expectedCount: 1,
			expectedText:  []string{"Task 1"},
			update:        "Single JSON line should be parsed correctly",
		},
		{
			name: "multiple valid JSON",
			fileContent: `{"text":"Task 1","created_at":"2024-01-01T10:00:00Z"}
{"text":"Task 2","created_at":"2024-01-01T11:00:00Z"}
{"text":"Task 3","created_at":"2024-01-01T12:00:00Z"}`,
			expectedCount: 3,
			expectedText:  []string{"Task 1", "Task 2", "Task 3"},
			update:        "Multiple JSON lines should be parsed correctly",
		},
		{
			name: "legacy plain text format",
			fileContent: `Plain text task 1
Plain text task 2`,
			expectedCount: 2,
			expectedText:  []string{"Plain text task 1", "Plain text task 2"},
			update:        "Legacy plain text format should be converted to Todo",
		},
		{
			name: "malformed JSON skipped",
			fileContent: `{"text":"Task 1","created_at":"2024-01-01T10:00:00Z"}
{invalid json}
{"text":"Task 2","created_at":"2024-01-01T11:00:00Z"}`,
			expectedCount: 3,
			expectedText:  []string{"Task 1", "{invalid json}", "Task 2"},
			update:        "Malformed JSON should be treated as plain text",
		},
		{
			name:          "UTF-8 characters in JSON",
			fileContent:   `{"text":"世界你好 🌍","created_at":"2024-01-01T10:00:00Z"}`,
			expectedCount: 1,
			expectedText:  []string{"世界你好 🌍"},
			update:        "UTF-8 characters should be preserved",
		},
		{
			name: "mixed JSON and plain text",
			fileContent: `{"text":"JSON Task","created_at":"2024-01-01T10:00:00Z"}
Plain text task
{"text":"Another JSON","created_at":"2024-01-01T11:00:00Z"}`,
			expectedCount: 3,
			expectedText:  []string{"JSON Task", "Plain text task", "Another JSON"},
			update:        "Mixed formats should be handled",
		},
		{
			name: "blank lines ignored",
			fileContent: `{"text":"Task 1","created_at":"2024-01-01T10:00:00Z"}

{"text":"Task 2","created_at":"2024-01-01T11:00:00Z"}

{"text":"Task 3","created_at":"2024-01-01T12:00:00Z"}`,
			expectedCount: 3,
			expectedText:  []string{"Task 1", "Task 2", "Task 3"},
			update:        "Blank lines should be ignored",
		},
		{
			name:          "JSON with updates array",
			fileContent:   `{"text":"Task with desc","updates":["desc1","desc2"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedCount: 1,
			expectedText:  []string{"Task with desc"},
			update:        "JSON with updates array should be parsed",
		},
		{
			name:          "JSON with legacy description string",
			fileContent:   `{"text":"Task with old desc","description":"old format","created_at":"2024-01-01T10:00:00Z"}`,
			expectedCount: 1,
			expectedText:  []string{"Task with old desc"},
			update:        "Old description format should be converted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_todos.txt")

			// Write test content to file
			if tt.fileContent != "" {
				err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}
			// If fileContent is empty, test non-existent file case

			// Load todos
			todos := loadTodos(tmpFile)

			// Verify count
			if len(todos) != tt.expectedCount {
				t.Errorf("Expected %d todos, got %d", tt.expectedCount, len(todos))
			}

			// Verify text content
			for i, expectedText := range tt.expectedText {
				if i >= len(todos) {
					t.Errorf("Missing todo at index %d", i)
					continue
				}
				if todos[i].Text != expectedText {
					t.Errorf("Todo[%d].Text = %q, want %q", i, todos[i].Text, expectedText)
				}
			}

			// Verify CreatedAt is set
			for i, todo := range todos {
				if todo.CreatedAt.IsZero() {
					t.Errorf("Todo[%d].CreatedAt should not be zero", i)
				}
			}
		})
	}
}

func TestLoadTodosNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.txt")

	todos := loadTodos(nonExistentFile)

	if len(todos) != 0 {
		t.Errorf("Expected empty slice for non-existent file, got %d todos", len(todos))
	}
}

func TestSaveTodos(t *testing.T) {
	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name          string
		todos         []Todo
		expectedLines int
		update        string
	}{
		{
			name:          "empty list",
			todos:         []Todo{},
			expectedLines: 0,
			update:        "Empty todo list should create empty file",
		},
		{
			name: "single todo",
			todos: []Todo{
				{Text: "Task 1", CreatedAt: now},
			},
			expectedLines: 1,
			update:        "Single todo should create one line",
		},
		{
			name: "multiple todos",
			todos: []Todo{
				{Text: "Task 1", CreatedAt: now},
				{Text: "Task 2", CreatedAt: now},
				{Text: "Task 3", CreatedAt: now},
			},
			expectedLines: 3,
			update:        "Multiple todos should create multiple lines",
		},
		{
			name: "todo with updates",
			todos: []Todo{
				{
					Text:      "Task with desc",
					Updates:   []string{"desc1", "desc2"},
					CreatedAt: now,
				},
			},
			expectedLines: 1,
			update:        "Todo with updates should be serialized",
		},
		{
			name: "completed todo",
			todos: []Todo{
				{
					Text:        "Completed task",
					CreatedAt:   now,
					CompletedAt: &completedTime,
				},
			},
			expectedLines: 1,
			update:        "Completed todo should preserve CompletedAt",
		},
		{
			name: "UTF-8 characters",
			todos: []Todo{
				{Text: "世界你好 🌍", CreatedAt: now},
			},
			expectedLines: 1,
			update:        "UTF-8 characters should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_save.txt")

			// Save todos
			err := saveTodos(tmpFile, tt.todos)
			if err != nil {
				t.Fatalf("saveTodos() error = %v", err)
			}

			// Load back and verify
			loaded := loadTodos(tmpFile)
			if len(loaded) != len(tt.todos) {
				t.Errorf("Loaded %d todos, expected %d", len(loaded), len(tt.todos))
			}

			// Verify each todo
			for i, original := range tt.todos {
				if i >= len(loaded) {
					t.Errorf("Missing todo at index %d", i)
					continue
				}

				if loaded[i].Text != original.Text {
					t.Errorf("Todo[%d].Text = %q, want %q", i, loaded[i].Text, original.Text)
				}

				if len(loaded[i].Updates) != len(original.Updates) {
					t.Errorf("Todo[%d].Updates length = %d, want %d", i, len(loaded[i].Updates), len(original.Updates))
				}

				// Verify CompletedAt if set
				if original.CompletedAt != nil {
					if loaded[i].CompletedAt == nil {
						t.Errorf("Todo[%d].CompletedAt should not be nil", i)
					}
				}
			}
		})
	}
}

func TestBackupCompletedTodos(t *testing.T) {
	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name        string
		todos       []Todo
		expectError bool
		update      string
	}{
		{
			name:        "empty list",
			todos:       []Todo{},
			expectError: false,
			update:      "Empty list should create backup file",
		},
		{
			name: "single todo",
			todos: []Todo{
				{
					Text:        "Completed task",
					CreatedAt:   now,
					CompletedAt: &completedTime,
				},
			},
			expectError: false,
			update:      "Single completed todo should be backed up",
		},
		{
			name: "multiple todos",
			todos: []Todo{
				{Text: "Task 1", CreatedAt: now, CompletedAt: &completedTime},
				{Text: "Task 2", CreatedAt: now, CompletedAt: &completedTime},
				{Text: "Task 3", CreatedAt: now, CompletedAt: &completedTime},
			},
			expectError: false,
			update:      "Multiple todos should be backed up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to temporary directory for backup file creation
			tmpDir := t.TempDir()
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)
			os.Chdir(tmpDir)

			// Create backup
			backupFile, err := backupCompletedTodos(tt.todos)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				// Verify filename format includes date and count
				dateStr := now.Format("2006-01-02")
				expectedPrefix := "todo_completed_backup_" + dateStr
				if len(backupFile) < len(expectedPrefix) {
					t.Errorf("Backup filename %q too short", backupFile)
				} else if backupFile[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("Backup filename should start with %q, got %q", expectedPrefix, backupFile)
				}

				// Verify file exists
				if _, err := os.Stat(backupFile); os.IsNotExist(err) {
					t.Errorf("Backup file %q was not created", backupFile)
				}

				// Verify content
				loaded := loadTodos(backupFile)
				if len(loaded) != len(tt.todos) {
					t.Errorf("Backup contains %d todos, expected %d", len(loaded), len(tt.todos))
				}

				// Verify todos match
				for i, original := range tt.todos {
					if i >= len(loaded) {
						continue
					}
					if loaded[i].Text != original.Text {
						t.Errorf("Backup todo[%d].Text = %q, want %q", i, loaded[i].Text, original.Text)
					}
				}

				// Clean up backup file
				os.Remove(backupFile)
			}
		})
	}
}

func TestBackupCompletedTodosFilenameFormat(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)
	todos := []Todo{
		{Text: "Task 1", CreatedAt: now, CompletedAt: &completedTime},
		{Text: "Task 2", CreatedAt: now, CompletedAt: &completedTime},
	}

	backupFile, err := backupCompletedTodos(todos)
	if err != nil {
		t.Fatalf("backupCompletedTodos() error = %v", err)
	}

	// Verify filename format: todo_completed_backup_YYYY-MM-DD_count.txt
	dateStr := now.Format("2006-01-02")
	expectedFilename := "todo_completed_backup_" + dateStr + "_2.txt"

	if backupFile != expectedFilename {
		t.Errorf("Backup filename = %q, want %q", backupFile, expectedFilename)
	}

	// Clean up
	os.Remove(backupFile)
}

func TestFindBackupFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	tests := []struct {
		name          string
		createFiles   []string
		expectedCount int
		update        string
	}{
		{
			name:          "no backup files",
			createFiles:   []string{},
			expectedCount: 0,
			update:        "Empty directory should return no backup files",
		},
		{
			name: "single backup file",
			createFiles: []string{
				"todo_completed_backup_2024-01-15_5.txt",
			},
			expectedCount: 1,
			update:        "Should find single backup file",
		},
		{
			name: "multiple backup files",
			createFiles: []string{
				"todo_completed_backup_2024-01-15_5.txt",
				"todo_completed_backup_2024-01-16_3.txt",
				"todo_completed_backup_2024-01-17_8.txt",
			},
			expectedCount: 3,
			update:        "Should find all backup files",
		},
		{
			name: "mixed files",
			createFiles: []string{
				"todo_completed_backup_2024-01-15_5.txt",
				"todo_backlog.txt",
				"todo_ready.txt",
				"todo_completed.txt",
				"other_file.txt",
			},
			expectedCount: 1,
			update:        "Should only find backup files, not other files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean directory
			matches, _ := filepath.Glob("*")
			for _, match := range matches {
				os.Remove(match)
			}

			// Create test files
			for _, filename := range tt.createFiles {
				err := os.WriteFile(filename, []byte("test"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %q: %v", filename, err)
				}
			}

			// Find backup files
			backupFiles, err := findBackupFiles()
			if err != nil {
				t.Fatalf("findBackupFiles() error = %v", err)
			}

			if len(backupFiles) != tt.expectedCount {
				t.Errorf("findBackupFiles() returned %d files, want %d", len(backupFiles), tt.expectedCount)
			}

			// Verify all returned files match the pattern
			for _, file := range backupFiles {
				matched, err := filepath.Match("todo_completed_backup_*.txt", file)
				if err != nil {
					t.Fatalf("filepath.Match error: %v", err)
				}
				if !matched {
					t.Errorf("findBackupFiles() returned non-backup file: %q", file)
				}
			}
		})
	}
}

func TestLoadAllCompletedTodos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()
	completedTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name               string
		mainTodos          []Todo
		backupFiles        map[string][]Todo
		expectedTotalCount int
		update             string
	}{
		{
			name:               "no todos anywhere",
			mainTodos:          []Todo{},
			backupFiles:        map[string][]Todo{},
			expectedTotalCount: 0,
			update:             "No todos should return empty slice",
		},
		{
			name: "only main file",
			mainTodos: []Todo{
				{Text: "Main task 1", CreatedAt: now, CompletedAt: &completedTime},
				{Text: "Main task 2", CreatedAt: now, CompletedAt: &completedTime},
			},
			backupFiles:        map[string][]Todo{},
			expectedTotalCount: 2,
			update:             "Should load only main file todos",
		},
		{
			name:      "only backup files",
			mainTodos: []Todo{},
			backupFiles: map[string][]Todo{
				"todo_completed_backup_2024-01-15_2.txt": {
					{Text: "Backup task 1", CreatedAt: now, CompletedAt: &completedTime},
					{Text: "Backup task 2", CreatedAt: now, CompletedAt: &completedTime},
				},
			},
			expectedTotalCount: 2,
			update:             "Should load only backup todos",
		},
		{
			name: "main and backup files",
			mainTodos: []Todo{
				{Text: "Main task", CreatedAt: now, CompletedAt: &completedTime},
			},
			backupFiles: map[string][]Todo{
				"todo_completed_backup_2024-01-15_2.txt": {
					{Text: "Backup task 1", CreatedAt: now, CompletedAt: &completedTime},
					{Text: "Backup task 2", CreatedAt: now, CompletedAt: &completedTime},
				},
			},
			expectedTotalCount: 3,
			update:             "Should combine main and backup todos",
		},
		{
			name: "multiple backup files",
			mainTodos: []Todo{
				{Text: "Main task", CreatedAt: now, CompletedAt: &completedTime},
			},
			backupFiles: map[string][]Todo{
				"todo_completed_backup_2024-01-15_2.txt": {
					{Text: "Backup1 task1", CreatedAt: now, CompletedAt: &completedTime},
					{Text: "Backup1 task2", CreatedAt: now, CompletedAt: &completedTime},
				},
				"todo_completed_backup_2024-01-16_1.txt": {
					{Text: "Backup2 task", CreatedAt: now, CompletedAt: &completedTime},
				},
			},
			expectedTotalCount: 4,
			update:             "Should combine all files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean directory
			matches, _ := filepath.Glob("*")
			for _, match := range matches {
				os.Remove(match)
			}

			// Create main completed file
			if len(tt.mainTodos) > 0 {
				err := saveTodos(completedFile, tt.mainTodos)
				if err != nil {
					t.Fatalf("Failed to save main todos: %v", err)
				}
			}

			// Create backup files
			for filename, todos := range tt.backupFiles {
				err := saveTodos(filename, todos)
				if err != nil {
					t.Fatalf("Failed to save backup file %q: %v", filename, err)
				}
			}

			// Load all completed todos
			allTodos := loadAllCompletedTodos()

			if len(allTodos) != tt.expectedTotalCount {
				t.Errorf("loadAllCompletedTodos() returned %d todos, want %d", len(allTodos), tt.expectedTotalCount)
			}

			// Verify that all todos from all sources are present
			// Create a map of all expected todos
			expectedTodos := make(map[string]bool)
			for _, todo := range tt.mainTodos {
				expectedTodos[todo.Text] = false
			}
			for _, todos := range tt.backupFiles {
				for _, todo := range todos {
					expectedTodos[todo.Text] = false
				}
			}

			// Mark todos as found
			for _, todo := range allTodos {
				if _, exists := expectedTodos[todo.Text]; exists {
					expectedTodos[todo.Text] = true
				}
			}

			// Check all todos were found
			for text, found := range expectedTodos {
				if !found {
					t.Errorf("Expected todo %q not found in loadAllCompletedTodos result", text)
				}
			}
		})
	}
}

func TestCreateBackups(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()

	tests := []struct {
		name       string
		setupFiles map[string][]Todo
		update     string
	}{
		{
			name:       "no existing files",
			setupFiles: map[string][]Todo{},
			update:     "Should create backup directory even with no files",
		},
		{
			name: "all three files exist",
			setupFiles: map[string][]Todo{
				backlogFile: {
					{Text: "Backlog task 1", CreatedAt: now},
					{Text: "Backlog task 2", CreatedAt: now},
				},
				readyFile: {
					{Text: "Ready task 1", CreatedAt: now},
				},
				completedFile: {
					{Text: "Completed task 1", CreatedAt: now},
				},
			},
			update: "Should backup all three files",
		},
		{
			name: "partial files exist",
			setupFiles: map[string][]Todo{
				readyFile: {
					{Text: "Ready task 1", CreatedAt: now},
				},
			},
			update: "Should backup only existing files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean directory
			matches, _ := filepath.Glob("*")
			for _, match := range matches {
				os.RemoveAll(match)
			}

			// Create test files
			for filename, todos := range tt.setupFiles {
				err := saveTodos(filename, todos)
				if err != nil {
					t.Fatalf("Failed to create test file %q: %v", filename, err)
				}
			}

			// Create backups
			err := createBackups()
			if err != nil {
				t.Fatalf("createBackups() error = %v", err)
			}

			// Verify backup directory exists
			if _, err := os.Stat("backup"); os.IsNotExist(err) {
				t.Error("backup directory was not created")
			}

			// Verify backup files exist and match original content
			for filename, originalTodos := range tt.setupFiles {
				backupPath := filepath.Join("backup", filename+".bak")

				// Check file exists
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("backup file %q was not created", backupPath)
					continue
				}

				// Load and verify content
				backupTodos := loadTodos(backupPath)
				if len(backupTodos) != len(originalTodos) {
					t.Errorf("backup file %q has %d todos, want %d", backupPath, len(backupTodos), len(originalTodos))
				}

				for i, original := range originalTodos {
					if i >= len(backupTodos) {
						continue
					}
					if backupTodos[i].Text != original.Text {
						t.Errorf("backup todo[%d].Text = %q, want %q", i, backupTodos[i].Text, original.Text)
					}
				}
			}
		})
	}
}

func TestCreateBackupsOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	now := time.Now()

	// Create initial file and backup
	initialTodos := []Todo{
		{Text: "Initial task", CreatedAt: now},
	}
	err := saveTodos(readyFile, initialTodos)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	err = createBackups()
	if err != nil {
		t.Fatalf("First createBackups() error = %v", err)
	}

	// Modify the original file
	updatedTodos := []Todo{
		{Text: "Updated task 1", CreatedAt: now},
		{Text: "Updated task 2", CreatedAt: now},
	}
	err = saveTodos(readyFile, updatedTodos)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Create backups again (should overwrite)
	err = createBackups()
	if err != nil {
		t.Fatalf("Second createBackups() error = %v", err)
	}

	// Verify backup has the updated content
	backupPath := filepath.Join("backup", readyFile+".bak")
	backupTodos := loadTodos(backupPath)

	if len(backupTodos) != len(updatedTodos) {
		t.Errorf("backup should have %d todos after overwrite, got %d", len(updatedTodos), len(backupTodos))
	}

	for i, updated := range updatedTodos {
		if i >= len(backupTodos) {
			continue
		}
		if backupTodos[i].Text != updated.Text {
			t.Errorf("backup todo[%d].Text = %q, want %q (overwrite failed)", i, backupTodos[i].Text, updated.Text)
		}
	}
}
