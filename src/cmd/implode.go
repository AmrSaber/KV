package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

// implodeCmd represents the implode command
var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Permanently delete all keys and history",
	Long: `Permanently delete all keys and their history from the store.

Warning: This action cannot be undone. Configuration settings are preserved.`,
	Example: `  # Delete all data
  kv implode`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ClearDB()
	},
}

func init() {
	rootCmd.AddCommand(implodeCmd)
}
