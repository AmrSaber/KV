package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unlockCmd = &cobra.Command{
	Use:     "unlock",
	Short:   "A brief description of your command",
	GroupID: "encryption",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unlock called")
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)
}
