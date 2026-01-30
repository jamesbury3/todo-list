package model

import (
	"fmt"
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

// DayGroup represents todos completed on a specific day
type DayGroup struct {
	Date  time.Time
	Todos []Todo
}

// WeekGroup represents a week with multiple days
type WeekGroup struct {
	WeekStart time.Time
	WeekEnd   time.Time
	Days      []DayGroup
}

// groupTodosByWeek groups all completed todos by week, then by day
func groupTodosByWeek(todos []Todo) []WeekGroup {
	if len(todos) == 0 {
		return []WeekGroup{}
	}

	// First, sort todos by completion time (oldest first for chronological display)
	sorted := make([]Todo, len(todos))
	copy(sorted, todos)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CompletedAt == nil {
			return false
		}
		if sorted[j].CompletedAt == nil {
			return true
		}
		return sorted[i].CompletedAt.Before(*sorted[j].CompletedAt)
	})

	// Group by day first
	dayMap := make(map[string]*DayGroup)
	var dayKeys []string

	for _, todo := range sorted {
		if todo.CompletedAt == nil {
			continue
		}

		// Get day key (YYYY-MM-DD)
		dayKey := todo.CompletedAt.Format("2006-01-02")

		if _, exists := dayMap[dayKey]; !exists {
			dayMap[dayKey] = &DayGroup{
				Date:  truncateToDay(*todo.CompletedAt),
				Todos: []Todo{},
			}
			dayKeys = append(dayKeys, dayKey)
		}
		dayMap[dayKey].Todos = append(dayMap[dayKey].Todos, todo)
	}

	// Now group days into weeks
	weekMap := make(map[string]*WeekGroup)
	var weekKeys []string

	for _, dayKey := range dayKeys {
		dayGroup := dayMap[dayKey]
		weekStart := getWeekStart(dayGroup.Date)
		weekEnd := weekStart.AddDate(0, 0, 6)
		weekKey := weekStart.Format("2006-01-02")

		if _, exists := weekMap[weekKey]; !exists {
			weekMap[weekKey] = &WeekGroup{
				WeekStart: weekStart,
				WeekEnd:   weekEnd,
				Days:      []DayGroup{},
			}
			weekKeys = append(weekKeys, weekKey)
		}
		weekMap[weekKey].Days = append(weekMap[weekKey].Days, *dayGroup)
	}

	// Convert map to slice, maintaining chronological order (most recent first)
	var weeks []WeekGroup
	for i := len(weekKeys) - 1; i >= 0; i-- {
		week := weekMap[weekKeys[i]]
		// Reverse days within each week to show most recent first
		reverseDays := make([]DayGroup, len(week.Days))
		for j := 0; j < len(week.Days); j++ {
			reverseDays[j] = week.Days[len(week.Days)-1-j]
		}
		week.Days = reverseDays
		weeks = append(weeks, *week)
	}

	return weeks
}

// getWeekStart returns the start of the week (Sunday) for the given date
func getWeekStart(t time.Time) time.Time {
	// Get the day of week (0 = Sunday, 1 = Monday, etc.)
	dayOfWeek := int(t.Weekday())
	// Subtract that many days to get to Sunday
	weekStart := t.AddDate(0, 0, -dayOfWeek)
	return truncateToDay(weekStart)
}

// truncateToDay returns a time truncated to the start of the day
func truncateToDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// formatWeekRange formats a week range as "Jan 15 - Jan 21"
func formatWeekRange(start, end time.Time) string {
	// If same month
	if start.Month() == end.Month() {
		return fmt.Sprintf("%s %d - %d", start.Format("Jan"), start.Day(), end.Day())
	}
	// Different months
	return fmt.Sprintf("%s %d - %s %d", start.Format("Jan"), start.Day(), end.Format("Jan"), end.Day())
}

// formatDayHeader formats a day header as "Monday, Jan 15"
func formatDayHeader(t time.Time) string {
	today := truncateToDay(time.Now())
	yesterday := today.AddDate(0, 0, -1)

	if t.Equal(today) {
		return fmt.Sprintf("Today (%s)", t.Format("Monday, Jan 2"))
	} else if t.Equal(yesterday) {
		return fmt.Sprintf("Yesterday (%s)", t.Format("Monday, Jan 2"))
	}
	return t.Format("Monday, Jan 2")
}
