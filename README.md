# Todo List TUI

### Introduction

This is a terminal-based todo list application built for keyboard navigation with vim-like commands.

<img width="753" height="269" alt="Screenshot 2026-01-27 at 9 49 25 PM" src="https://github.com/user-attachments/assets/d4341861-fd9c-42ed-ad50-c73b49615bf3" />

### Getting Started

1. Install Go v1.25+ 
2. Clone this repository
3. Run the following command
```
go build .
```
4. Move the binary to the folder where you want to save your todo list
5. Run the binary and you're done!

### Usage

The todo list is split into 3 pages: Backlog, Ready, and Completed. Each of these has slightly different commands, but navigation is common throughout.

**Common**
- `j`/`k` - Navigate down/up
- `h`/`l` - Switch between views
- `d` - Delete todo
- `e` - Edit description
- `n` - Rename todo
- `i` - Toggle description
- `I` - Toggle all descriptions
- `q` - Quit

**Backlog**
- `a` - Add new todo
- `r` - Move to ready
- `J`/`K` - Reorder todos

**Ready**
- `a` - Add new todo
- `b` - Move to backlog
- `x` - Mark as complete
- `J`/`K` - Reorder todos

**Completed**
- `u` - Undo completion
- `B` - Empty all completed todos into a backup text file named after the current date and number of completed todos

### Additional Notes

- The Completed page will only show the 10 most recently completed todos. To see the rest, you can always open todo_completed.txt and your backup files.
