package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list history called")
	},
}

func init() {
	historyCmd.AddCommand(historyListCmd)
}
