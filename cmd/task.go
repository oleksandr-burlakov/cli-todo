package cmd

import (
	"fmt"
	"time"

	"github.com/cli-todo/internal/store"
	"github.com/spf13/cobra"
)

var (
	taskWorkspace string
	taskProject   string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a task",
	Long:  "Create a task in a workspace. Use --project to put it in a project/list, or omit for default list.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, taskWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q: %w", taskWorkspace, err)
		}
		var projectID *int64
		if taskProject != "" {
			projects, err := store.ListProjects(db, w.ID)
			if err != nil {
				return err
			}
			for _, p := range projects {
				if p.Name == taskProject {
					projectID = &p.ID
					break
				}
			}
			if projectID == nil {
				return fmt.Errorf("project %q not found in workspace %q", taskProject, taskWorkspace)
			}
		}
		due := parseDue(dueDate)
		t, err := store.CreateTask(db, w.ID, projectID, args[0], description, status, priority, due)
		if err != nil {
			return err
		}
		fmt.Printf("Created task %d: %s [%s]\n", t.ID, t.Title, t.Status)
		return nil
	},
}

var (
	description string
	status      string
	priority    string
	dueDate     string
)

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  "List tasks in a workspace. Use --project to filter by project, or omit for default list.",
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, taskWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q: %w", taskWorkspace, err)
		}
		var projectID *int64
		if taskProject != "" {
			projects, err := store.ListProjects(db, w.ID)
			if err != nil {
				return err
			}
			for _, p := range projects {
				if p.Name == taskProject {
					projectID = &p.ID
					break
				}
			}
			if projectID == nil {
				return fmt.Errorf("project %q not found", taskProject)
			}
		}
		list, err := store.ListTasks(db, w.ID, projectID)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("No tasks.")
			return nil
		}
		for _, t := range list {
			due := ""
			if t.DueDate != nil {
				due = " due:" + t.DueDate.Format("2006-01-02")
			}
			pri := ""
			if t.Priority != "" {
				pri = " [" + t.Priority + "]"
			}
			fmt.Printf("  %d  [%s]%s  %s%s\n", t.ID, t.Status, pri, t.Title, due)
		}
		return nil
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("task id must be a number")
		}
		t, err := store.GetTask(db, id)
		if err != nil {
			return err
		}
		title := t.Title
		if editTitle != "" {
			title = editTitle
		}
		desc := t.Description
		if editDescription != "" {
			desc = editDescription
		}
		st := t.Status
		if editStatus != "" {
			st = editStatus
		}
		pri := t.Priority
		if editPriority != "" {
			pri = editPriority
		}
		due := t.DueDate
		if editDue != "" {
			due = parseDue(editDue)
		}
		_, err = store.UpdateTask(db, id, title, desc, st, pri, due)
		if err != nil {
			return err
		}
		fmt.Printf("Updated task %d\n", id)
		return nil
	},
}

var (
	editTitle       string
	editDescription string
	editStatus      string
	editPriority    string
	editDue         string
)

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("task id must be a number")
		}
		if err := store.DeleteTask(db, id); err != nil {
			return err
		}
		fmt.Printf("Deleted task %d\n", id)
		return nil
	},
}

func parseDue(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339[:10]} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return &t
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.PersistentFlags().StringVarP(&taskWorkspace, "workspace", "w", "", "Workspace name (required)")
	taskCmd.PersistentFlags().StringVarP(&taskProject, "project", "p", "", "Project/list name (optional; default list if omitted)")
	taskCmd.MarkPersistentFlagRequired("workspace")

	taskCreateCmd.Flags().StringVarP(&description, "description", "d", "", "Task description")
	taskCreateCmd.Flags().StringVarP(&status, "status", "s", "todo", "Status: todo, in_progress, done")
	taskCreateCmd.Flags().StringVarP(&priority, "priority", "", "", "Priority: low, medium, high")
	taskCreateCmd.Flags().StringVar(&dueDate, "due", "", "Due date (YYYY-MM-DD)")

	taskEditCmd.Flags().StringVar(&editTitle, "title", "", "New title")
	taskEditCmd.Flags().StringVar(&editDescription, "description", "", "New description")
	taskEditCmd.Flags().StringVar(&editStatus, "status", "", "New status: todo, in_progress, done")
	taskEditCmd.Flags().StringVar(&editPriority, "priority", "", "New priority: low, medium, high")
	taskEditCmd.Flags().StringVar(&editDue, "due", "", "New due date (YYYY-MM-DD)")

	taskCmd.AddCommand(taskCreateCmd, taskListCmd, taskEditCmd, taskDeleteCmd)
}
