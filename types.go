package main

import "time"

const (
	backlogFile    = "todo_backlog.txt"
	inProgressFile = "todo_inprogress.txt"
	completedFile  = "todo_completed.txt"
)

type view int

const (
	viewBacklog view = iota
	viewInProgress
	viewCompleted
)

type Todo struct {
	Text        string     `json:"text"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type model struct {
	backlog                []Todo
	inProgress             []Todo
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
	confirmingDelete       bool
	message                string
	textInputCursor        int // Cursor position within text input fields (for arrow key navigation)
}
