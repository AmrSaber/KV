package cmd

import (
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [command]",
	Short: "Histroy-related commands",
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
