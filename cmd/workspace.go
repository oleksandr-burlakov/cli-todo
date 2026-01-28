package cmd

import (
	"fmt"

	"github.com/cli-todo/internal/store"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces",
}

var workspaceCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.CreateWorkspace(db, args[0])
		if err != nil {
			fmt.Printf("Error %s", err.Error())
			return err
		}
		fmt.Printf("Created workspace %q (id %d)\n", w.Name, w.ID)
		return nil
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := store.ListWorkspaces(db)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("No workspaces. Create one with: todo workspace create <name>")
			return nil
		}
		for _, w := range list {
			fmt.Printf("  %d  %s\n", w.ID, w.Name)
		}
		return nil
	},
}

var workspaceDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a workspace and all its projects and tasks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := store.GetWorkspaceByName(db, args[0])
		if err != nil {
			return fmt.Errorf("workspace %q: %w", args[0], err)
		}
		if err := store.DeleteWorkspace(db, w.ID); err != nil {
			return err
		}
		fmt.Printf("Deleted workspace %q\n", w.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceCreateCmd, workspaceListCmd, workspaceDeleteCmd)
}
