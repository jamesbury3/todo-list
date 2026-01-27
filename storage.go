package main

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
