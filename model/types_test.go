package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTodoUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name                string
		jsonData            string
		expectedText        string
		expectedUpdateCount int
		expectedUpdates     []string
		expectError         bool
		update              string
	}{
		{
			name:                "new format - updates as array",
			jsonData:            `{"text":"Task 1","updates":["desc1","desc2"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 1",
			expectedUpdateCount: 2,
			expectedUpdates:     []string{"desc1", "desc2"},
			expectError:         false,
			update:              "New format with []string updates",
		},
		{
			name:                "legacy description as string",
			jsonData:            `{"text":"Task 2","description":"old desc","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 2",
			expectedUpdateCount: 1,
			expectedUpdates:     []string{"old desc"},
			expectError:         false,
			update:              "Legacy description string should convert to []string",
		},
		{
			name:                "empty update string",
			jsonData:            `{"text":"Task 3","description":"","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 3",
			expectedUpdateCount: 0,
			expectedUpdates:     []string{},
			expectError:         false,
			update:              "Empty update string should result in nil slice",
		},
		{
			name:                "no updates field",
			jsonData:            `{"text":"Task 4","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 4",
			expectedUpdateCount: 0,
			expectedUpdates:     []string{},
			expectError:         false,
			update:              "Missing updates field should be nil",
		},
		{
			name:                "empty updates array",
			jsonData:            `{"text":"Task 5","updates":[],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 5",
			expectedUpdateCount: 0,
			expectedUpdates:     []string{},
			expectError:         false,
			update:              "Empty updates array",
		},
		{
			name:                "single item updates array",
			jsonData:            `{"text":"Task 6","updates":["only one"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 6",
			expectedUpdateCount: 1,
			expectedUpdates:     []string{"only one"},
			expectError:         false,
			update:              "Single item array",
		},
		{
			name:                "UTF-8 in updates",
			jsonData:            `{"text":"Task 7","updates":["世界","🌍"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:        "Task 7",
			expectedUpdateCount: 2,
			expectedUpdates:     []string{"世界", "🌍"},
			expectError:         false,
			update:              "UTF-8 characters in updates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var todo Todo
			err := json.Unmarshal([]byte(tt.jsonData), &todo)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				// Verify text
				if todo.Text != tt.expectedText {
					t.Errorf("Text = %q, want %q", todo.Text, tt.expectedText)
				}

				// Verify update count
				if len(todo.Updates) != tt.expectedUpdateCount {
					t.Errorf("Updates count = %d, want %d", len(todo.Updates), tt.expectedUpdateCount)
				}

				// Verify update content
				for i, expectedUpdate := range tt.expectedUpdates {
					if i >= len(todo.Updates) {
						t.Errorf("Missing update at index %d", i)
						continue
					}
					if todo.Updates[i] != expectedUpdate {
						t.Errorf("Updates[%d] = %q, want %q", i, todo.Updates[i], expectedUpdate)
					}
				}

				// Verify CreatedAt is parsed
				if todo.CreatedAt.IsZero() {
					t.Error("CreatedAt should not be zero")
				}
			}
		})
	}
}

func TestTodoUnmarshalJSONWithCompletedAt(t *testing.T) {
	jsonData := `{"text":"Completed Task","created_at":"2024-01-01T10:00:00Z","completed_at":"2024-01-01T11:00:00Z"}`

	var todo Todo
	err := json.Unmarshal([]byte(jsonData), &todo)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if todo.Text != "Completed Task" {
		t.Errorf("Text = %q, want 'Completed Task'", todo.Text)
	}

	if todo.CompletedAt == nil {
		t.Fatal("CompletedAt should not be nil")
	}

	// Verify CompletedAt is after CreatedAt
	if !todo.CompletedAt.After(todo.CreatedAt) {
		t.Error("CompletedAt should be after CreatedAt")
	}
}

func TestTodoMarshalJSON(t *testing.T) {
	now := time.Now()
	completedTime := now.Add(1 * time.Hour)

	tests := []struct {
		name        string
		todo        Todo
		checkFields []string
		update      string
	}{
		{
			name: "todo with updates array",
			todo: Todo{
				Text:      "Task 1",
				Updates:   []string{"desc1", "desc2"},
				CreatedAt: now,
			},
			checkFields: []string{`"text":"Task 1"`, `"updates"`, `"created_at"`},
			update:      "Should marshal updates as array",
		},
		{
			name: "todo without updates",
			todo: Todo{
				Text:      "Task 2",
				CreatedAt: now,
			},
			checkFields: []string{`"text":"Task 2"`, `"created_at"`},
			update:      "Should omit empty updates",
		},
		{
			name: "completed todo",
			todo: Todo{
				Text:        "Completed",
				CreatedAt:   now,
				CompletedAt: &completedTime,
			},
			checkFields: []string{`"text":"Completed"`, `"completed_at"`},
			update:      "Should include completed_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.todo)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			jsonStr := string(data)

			// Check for expected fields
			for _, field := range tt.checkFields {
				if !contains(jsonStr, field) {
					t.Errorf("JSON should contain %q, got: %s", field, jsonStr)
				}
			}

			// Verify can unmarshal back
			var unmarshaled Todo
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if unmarshaled.Text != tt.todo.Text {
				t.Errorf("Round-trip Text = %q, want %q", unmarshaled.Text, tt.todo.Text)
			}
		})
	}
}

func TestTodoBackwardCompatibility(t *testing.T) {
	// Test that we can read old format and write new format
	oldFormat := `{"text":"Old Task","description":"single string","created_at":"2024-01-01T10:00:00Z"}`

	var todo Todo
	err := json.Unmarshal([]byte(oldFormat), &todo)
	if err != nil {
		t.Fatalf("Failed to unmarshal old format: %v", err)
	}

	// Verify conversion
	if len(todo.Updates) != 1 {
		t.Errorf("Updates should be converted to array with 1 element, got %d", len(todo.Updates))
	}
	if len(todo.Updates) > 0 && todo.Updates[0] != "single string" {
		t.Errorf("Updates[0] = %q, want 'single string'", todo.Updates[0])
	}

	// Marshal back to JSON (should use new format)
	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	// New format should have array
	if !contains(jsonStr, `"updates":["single string"]`) {
		t.Errorf("Marshaled JSON should use array format, got: %s", jsonStr)
	}
}
