package cmd

import (
	"database/sql"
	"fmt"

	"github.com/cli-todo/internal/store"
	"github.com/spf13/cobra"
)

var dbPath string
var db *sql.DB

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "CLI todo app with workspaces and projects",
	Long:  "Track tasks in workspaces (personal, work, daily, etc.) and optional projects/lists. Data stored locally in SQLite.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		db, err = store.Open(dbPath)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Path to SQLite database (default: config dir/cli-todo/todo.db)")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
