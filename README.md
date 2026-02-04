# Todo List TUI

### Introduction

This is a terminal-based todo list application built for keyboard navigation with vim-like commands.

<img width="610" height="214" alt="Screenshot 2026-02-02 213526" src="https://github.com/user-attachments/assets/c7dd0c93-3dce-4aeb-8432-680afb927061" />

### Installation

1. Download the most recent binary from [Releases](https://github.com/jamesbury3/Todo-List-TUI/releases)
2. Place the binary in the location where you want your todos to be saved, such as a cloud synced folder
3. Run the application, and you're good to go!
   - You can also copy the binary into multiple folders if you prefer to have separate todo lists for separate projects
   - Unless you're storing your todo lists on your Desktop, you may want to create a Desktop shortcut to the binary

### Usage

The todo list is split into 3 pages: Backlog, Ready, and Completed. Each of these has slightly different commands, but navigation is common throughout.

**Common**
- `j`/`k` - Navigate down/up
- `g`/`G` - Go to top/bottom of current list
- `h`/`l` - Switch between views
- `d` - Delete todo
- `u` - Add update
- `n` - Rename todo / edit update
- `i` - Toggle updates
- `I` - Toggle all updates
- `q` - Quit

**Backlog**
- `a` - Add new todo
- `A` - Add new todo to top
- `r` - Move to ready
- `J`/`K` - Reorder todos
- `t` - Move todo to top

**Ready**
- `a` - Add new todo
- `A` - Add new todo to top
- `b` - Move to backlog
- `x` - Mark as complete
- `J`/`K` - Reorder todos
- `t` - Move todo to top

**Completed**
- `r` - Move back to ready
- `p` - Toggle prettify view (shows all todos grouped by week/day)
- `P` - Export markdown (creates a markdown file with all todos including backups)
- `B` - Empty all completed todos into a backup text file named after the current date and number of completed todos

### Additional Notes

- The Completed page will only show the 10 most recently completed todos. To see the rest, you can always open todo_completed.txt and your backup files.

### Build Yourself

To build the application yourself with Go:

1. Install Go v1.25+ 
2. Clone this repository
3. Run the following command
```
go build
```
4. Move the binary to the folder where you want to save your todo list
5. Run the binary and you're done!
