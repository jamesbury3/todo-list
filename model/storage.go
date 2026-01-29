package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func loadTodos(filename string) []Todo {
	file, err := os.Open(filename)
	if err != nil {
		return []Todo{}
	}
	defer file.Close()

	var todos []Todo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			var todo Todo
			if err := json.Unmarshal([]byte(line), &todo); err != nil {
				// Handle old format - plain text
				todo = Todo{
					Text:      line,
					CreatedAt: time.Now(),
				}
			}
			todos = append(todos, todo)
		}
	}
	return todos
}

func saveTodos(filename string, todos []Todo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, todo := range todos {
		data, err := json.Marshal(todo)
		if err != nil {
			continue
		}
		fmt.Fprintln(writer, string(data))
	}
	return writer.Flush()
}

func backupCompletedTodos(todos []Todo) (string, error) {
	// Generate backup filename with current date and number of todos
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	filename := fmt.Sprintf("todo_completed_backup_%s_%d.txt", dateStr, len(todos))

	// Save todos to backup file
	if err := saveTodos(filename, todos); err != nil {
		return "", err
	}

	return filename, nil
}
