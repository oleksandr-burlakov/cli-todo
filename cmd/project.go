package cmd

import (
	"fmt"

	"github.com/cli-todo/internal/store"
	"github.com/spf13/cobra"
)

var projectWorkspace string

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects/lists within a workspace",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a project in a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, projectWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q: %w", projectWorkspace, err)
		}
		p, err := store.CreateProject(db, w.ID, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Created project %q in %s (id %d)\n", p.Name, projectWorkspace, p.ID)
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in a workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, projectWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q: %w", projectWorkspace, err)
		}
		list, err := store.ListProjects(db, w.ID)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Printf("No projects in %q. Tasks without a project go to the default list.\n", projectWorkspace)
			return nil
		}
		for _, p := range list {
			fmt.Printf("  %d  %s\n", p.ID, p.Name)
		}
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a project (tasks move to default list)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, projectWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q: %w", projectWorkspace, err)
		}
		projects, err := store.ListProjects(db, w.ID)
		if err != nil {
			return err
		}
		var targetID int64 = -1
		for _, p := range projects {
			if p.Name == args[0] {
				targetID = p.ID
				break
			}
		}
		if targetID < 0 {
			return fmt.Errorf("project %q not found in workspace %q", args[0], projectWorkspace)
		}
		if err := store.DeleteProject(db, targetID); err != nil {
			return err
		}
		fmt.Printf("Deleted project %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.PersistentFlags().StringVarP(&projectWorkspace, "workspace", "w", "", "Workspace name (required)")
	projectCmd.MarkPersistentFlagRequired("workspace")
	projectCmd.AddCommand(projectCreateCmd, projectListCmd, projectDeleteCmd)
}
