package main

import (
	"fmt"
	"os"

	"github.com/cli-todo/cmd"
	"github.com/cli-todo/internal/store"
	"github.com/cli-todo/internal/tui"
)

func main() {
	// No args: run interactive TUI. With args: run CLI (e.g. todo workspace list).
	if len(os.Args) == 1 {
		db, err := store.Open("")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Database:", err)
			os.Exit(1)
		}
		defer db.Close()
		if err := tui.Run(db); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
