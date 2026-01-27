package main

import "time"

const (
	inProgressFile = "todo_inprogress.txt"
	completedFile  = "todo_completed.txt"
)

type view int

const (
	viewInProgress view = iota
	viewCompleted
)

type Todo struct {
	Text        string     `json:"text"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type model struct {
	inProgress         []Todo
	completed          []Todo
	displayedCompleted []Todo // Stores the filtered/sorted completed todos for display
	cursor             int
	currentView        view
	adding             bool
	newTodo            string
	message            string
}
