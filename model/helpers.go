package model

import (
	"sort"
	"time"
)

// getCurrentList returns the list of todos for the current view
func (m *Model) getCurrentList() []Todo {
	switch m.currentView {
	case viewBacklog:
		return m.backlog
	case viewReady:
		return m.ready
	default:
		return m.displayedCompleted
	}
}

// updateCompletedTodo finds and updates a todo in the completed list
func (m *Model) updateCompletedTodo(updateFn func(*Todo)) {
	if m.cursor >= len(m.displayedCompleted) {
		return
	}
	todoToUpdate := m.displayedCompleted[m.cursor]
	for i := range m.completed {
		if m.completed[i].Text == todoToUpdate.Text && m.completed[i].CreatedAt.Equal(todoToUpdate.CreatedAt) {
			updateFn(&m.completed[i])
			break
		}
	}
	m.updateDisplayedCompleted()
}

// swapTodos swaps two adjacent items in a list and saves
func swapTodos(list []Todo, idx1, idx2 int, filename string) {
	list[idx1], list[idx2] = list[idx2], list[idx1]
	saveTodos(filename, list)
}

// updateDisplayedCompleted sorts and limits the completed list for display
func (m *Model) updateDisplayedCompleted() {
	if len(m.completed) == 0 {
		m.displayedCompleted = []Todo{}
		return
	}

	// Sort by completion time (most recent first)
	sorted := make([]Todo, len(m.completed))
	copy(sorted, m.completed)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CompletedAt == nil {
			return false
		}
		if sorted[j].CompletedAt == nil {
			return true
		}
		return sorted[i].CompletedAt.After(*sorted[j].CompletedAt)
	})

	// Take only the first 10
	if len(sorted) > 10 {
		m.displayedCompleted = sorted[:10]
	} else {
		m.displayedCompleted = sorted
	}
}

// countCompletedToday returns the number of todos completed today
func (m *Model) countCompletedToday() int {
	today := time.Now()
	todayYear, todayMonth, todayDay := today.Date()
	count := 0

	for _, todo := range m.completed {
		if todo.CompletedAt != nil {
			completedYear, completedMonth, completedDay := todo.CompletedAt.Date()
			if todayYear == completedYear && todayMonth == completedMonth && todayDay == completedDay {
				count++
			}
		}
	}

	return count
}
