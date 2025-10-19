package cmd

import (
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history [command]",
	Aliases: []string{"h"},
	Short:   "Histroy-related commands",
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
