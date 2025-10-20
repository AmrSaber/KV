package cmd

import (
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:     "history [command]",
	Aliases: []string{"h"},
	Short:   "Version control commands for key history",
	Long:    `Version control commands for viewing, reverting, and managing key history.`,
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
