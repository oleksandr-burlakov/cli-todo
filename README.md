# CLI Todo

A local CLI todo app with **workspaces** and **projects/lists**. Data is stored in SQLite (single file, cross-platform).

## Interactive TUI

Run **with no arguments** to start the interactive app:

```bash
./todo
# or on Windows: .\todo.exe
```

- **↑/↓** — move, **Enter** — open workspace/list or select
- **a** — add (workspace, project, or task)
- **d** — delete selected
- **← / Backspace** — go back
- **q** — quit  
- Type to filter lists

Subcommands (e.g. `./todo workspace list`) still work for scripting.

## Features

- **Workspaces** — e.g. `personal`, `family`, `daily`, `work`
- **Projects/lists** (optional) — inside a workspace; e.g. "books to read", "groceries". If you don't set a project, the task goes to the **default list** for that workspace.
- **Tasks** — title, optional description, status (`todo` / `in_progress` / `done`), optional priority (`low` / `medium` / `high`), optional due date

## Requirements

- Go 1.22+ (install from [go.dev](https://go.dev/dl/) or your package manager)

## Build

**Linux / macOS:**

```bash
go mod tidy
go build -o todo .
```

**Windows (PowerShell):**

```powershell
.\build.ps1
# or with custom output: .\build.ps1 -Output ".\bin\todo.exe"
```

Cross-compile (e.g. Windows from Linux):

```bash
GOOS=windows GOARCH=amd64 go build -o todo.exe .
```

## Data location

By default the SQLite database is **`todo.db` in the current working directory** (the folder from which you run the app). So if you run `./todo` from your project folder, the file appears there.

Override with `--db /path/to/todo.db` when using the CLI.

## Usage

### Workspaces

```bash
# Create workspaces
./todo workspace create personal
./todo workspace create work
./todo workspace create daily

# List
./todo workspace list

# Delete (removes all projects and tasks in it)
./todo workspace delete work
```

### Projects (lists inside a workspace)

```bash
# Create projects in a workspace
./todo project create "books to read" --workspace personal
./todo project create "groceries" --workspace personal
./todo project list --workspace personal

# Delete (tasks in that project move to the default list)
./todo project delete "groceries" --workspace personal
```

### Tasks

```bash
# Create in default list (no project)
./todo task create "Call mom" --workspace personal

# Create in a project
./todo task create "Buy milk" --workspace personal --project groceries
./todo task create "Read chapter 1" --workspace personal --project "books to read" --due 2026-02-01 --priority high

# List (default list only)
./todo task list --workspace personal

# List tasks in a project
./todo task list --workspace personal --project groceries

# Edit
./todo task edit 1 --status in_progress
./todo task edit 1 --title "Call mom (birthday)" --due 2026-01-30

# Delete
./todo task delete 1
```

## Task fields

| Field         | Required | Values / format                          |
|---------------|----------|------------------------------------------|
| title         | yes      | any                                      |
| description   | no       | any                                      |
| status        | no       | `todo`, `in_progress`, `done` (default: todo) |
| priority      | no       | `low`, `medium`, `high`                  |
| due date      | no       | `YYYY-MM-DD`                             |

## Future ideas (not implemented yet)

- **Filters and queryable boards** — e.g. "all done tasks", "planned for today" in one workspace or across workspaces. The SQLite schema (indexes on `status`, `due_date`, `workspace_id`) is ready for this; you can add subcommands that run custom queries later.

## Tech

- **Language**: Go
- **Storage**: SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)
- **CLI**: [Cobra](https://github.com/spf13/cobra)
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles)
