package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTodoUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name                 string
		jsonData             string
		expectedText         string
		expectedDescCount    int
		expectedDescriptions []string
		expectError          bool
		description          string
	}{
		{
			name:                 "new format - description as array",
			jsonData:             `{"text":"Task 1","description":["desc1","desc2"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 1",
			expectedDescCount:    2,
			expectedDescriptions: []string{"desc1", "desc2"},
			expectError:          false,
			description:          "New format with []string description",
		},
		{
			name:                 "old format - description as string",
			jsonData:             `{"text":"Task 2","description":"old desc","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 2",
			expectedDescCount:    1,
			expectedDescriptions: []string{"old desc"},
			expectError:          false,
			description:          "Old format with string description should convert to []string",
		},
		{
			name:                 "empty description string",
			jsonData:             `{"text":"Task 3","description":"","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 3",
			expectedDescCount:    0,
			expectedDescriptions: []string{},
			expectError:          false,
			description:          "Empty description string should result in nil slice",
		},
		{
			name:                 "no description field",
			jsonData:             `{"text":"Task 4","created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 4",
			expectedDescCount:    0,
			expectedDescriptions: []string{},
			expectError:          false,
			description:          "Missing description field should be nil",
		},
		{
			name:                 "empty description array",
			jsonData:             `{"text":"Task 5","description":[],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 5",
			expectedDescCount:    0,
			expectedDescriptions: []string{},
			expectError:          false,
			description:          "Empty description array",
		},
		{
			name:                 "single item description array",
			jsonData:             `{"text":"Task 6","description":["only one"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 6",
			expectedDescCount:    1,
			expectedDescriptions: []string{"only one"},
			expectError:          false,
			description:          "Single item array",
		},
		{
			name:                 "UTF-8 in description",
			jsonData:             `{"text":"Task 7","description":["世界","🌍"],"created_at":"2024-01-01T10:00:00Z"}`,
			expectedText:         "Task 7",
			expectedDescCount:    2,
			expectedDescriptions: []string{"世界", "🌍"},
			expectError:          false,
			description:          "UTF-8 characters in descriptions",
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

				// Verify description count
				if len(todo.Description) != tt.expectedDescCount {
					t.Errorf("Description count = %d, want %d", len(todo.Description), tt.expectedDescCount)
				}

				// Verify description content
				for i, expectedDesc := range tt.expectedDescriptions {
					if i >= len(todo.Description) {
						t.Errorf("Missing description at index %d", i)
						continue
					}
					if todo.Description[i] != expectedDesc {
						t.Errorf("Description[%d] = %q, want %q", i, todo.Description[i], expectedDesc)
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
		description string
	}{
		{
			name: "todo with description array",
			todo: Todo{
				Text:        "Task 1",
				Description: []string{"desc1", "desc2"},
				CreatedAt:   now,
			},
			checkFields: []string{`"text":"Task 1"`, `"description"`, `"created_at"`},
			description: "Should marshal description as array",
		},
		{
			name: "todo without description",
			todo: Todo{
				Text:      "Task 2",
				CreatedAt: now,
			},
			checkFields: []string{`"text":"Task 2"`, `"created_at"`},
			description: "Should omit empty description",
		},
		{
			name: "completed todo",
			todo: Todo{
				Text:        "Completed",
				CreatedAt:   now,
				CompletedAt: &completedTime,
			},
			checkFields: []string{`"text":"Completed"`, `"completed_at"`},
			description: "Should include completed_at",
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
	if len(todo.Description) != 1 {
		t.Errorf("Description should be converted to array with 1 element, got %d", len(todo.Description))
	}
	if len(todo.Description) > 0 && todo.Description[0] != "single string" {
		t.Errorf("Description[0] = %q, want 'single string'", todo.Description[0])
	}

	// Marshal back to JSON (should use new format)
	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	// New format should have array
	if !contains(jsonStr, `"description":["single string"]`) {
		t.Errorf("Marshaled JSON should use array format, got: %s", jsonStr)
	}
}
