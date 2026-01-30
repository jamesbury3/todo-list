package model

import (
	"encoding/json"
	"time"
)

const (
	backlogFile   = "todo_backlog.txt"
	readyFile     = "todo_ready.txt"
	completedFile = "todo_completed.txt"
)

type view int

const (
	viewBacklog view = iota
	viewReady
	viewCompleted
)

type Todo struct {
	Text        string     `json:"text"`
	Description []string   `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// UnmarshalJSON provides backward compatibility for loading old single-string descriptions
func (t *Todo) UnmarshalJSON(data []byte) error {
	type Alias Todo
	aux := &struct {
		Description interface{} `json:"description,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle description field - can be string (old format) or []string (new format)
	switch desc := aux.Description.(type) {
	case string:
		if desc != "" {
			t.Description = []string{desc}
		}
	case []interface{}:
		for _, v := range desc {
			if s, ok := v.(string); ok {
				t.Description = append(t.Description, s)
			}
		}
	}

	return nil
}

type Model struct {
	backlog                []Todo
	ready                  []Todo
	completed              []Todo
	displayedCompleted     []Todo // Stores the filtered/sorted completed todos for display
	cursor                 int
	currentView            view
	adding                 bool
	newTodo                string
	editingDescription     bool
	newDescription         string
	renamingTodo           bool
	newTodoName            string
	showingDescription     bool
	showingAllDescriptions bool
	showingCommands        bool
	confirmingDelete       bool
	navigatingDescriptions bool // True when in description navigation mode
	descriptionCursor      int  // Which description is selected (0-indexed)
	confirmingDeleteDesc   bool // True when confirming description deletion
	showingPrettify        bool // True when in prettify view (Completed tab only)
	message                string
	textInputCursor        int // Cursor position within text input fields (for arrow key navigation)
	width                  int // Terminal width
	height                 int // Terminal height
}
