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
	Text         string     `json:"text"`
	CompleteNote string     `json:"complete_note,omitempty"`
	Updates      []string   `json:"updates,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// UnmarshalJSON provides backward compatibility for loading old single-string descriptions.
func (t *Todo) UnmarshalJSON(data []byte) error {
	type Alias Todo
	aux := &struct {
		Updates     interface{} `json:"updates,omitempty"`
		Description interface{} `json:"description,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parseUpdates := func(raw interface{}) []string {
		var updates []string
		switch value := raw.(type) {
		case string:
			if value != "" {
				updates = []string{value}
			}
		case []interface{}:
			for _, v := range value {
				if s, ok := v.(string); ok {
					updates = append(updates, s)
				}
			}
		case []string:
			updates = append(updates, value...)
		}
		return updates
	}

	// Handle updates field - can be string (old format) or []string (new format).
	if aux.Updates != nil {
		t.Updates = parseUpdates(aux.Updates)
	} else if aux.Description != nil {
		// Legacy "description" field support.
		t.Updates = parseUpdates(aux.Description)
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
	addingToTop            bool // True when adding to top with 'A'
	newTodo                string
	editingUpdate          bool
	newUpdate              string
	editingCompleteNote    bool   // True when editing a complete note
	newCompleteNote        string // Buffer for complete note editing
	renamingTodo           bool
	newTodoName            string
	showingUpdate          bool
	showingAllUpdates      bool
	showingCommands        bool
	confirmingDelete       bool
	navigatingUpdates      bool // True when in update navigation mode
	updateCursor           int  // Which update is selected (0-indexed)
	confirmingDeleteUpdate bool // True when confirming update deletion
	showingPrettify        bool // True when in prettify view (Completed tab only)
	saveError              string
	message                string
	textInputCursor        int // Cursor position within text input fields (for arrow key navigation)
	width                  int // Terminal width
	height                 int // Terminal height
}
