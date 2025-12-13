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

This command is not suitable to run concurrently with other commands,
it is your responsibility to make sure no other KV command is being executed at the time of executing this command.
Running this command concurrently with another command can result in undefined behaviour.

This action cannot be undone. Configuration settings are preserved.`,
	Example: `  # Delete all data
  kv implode`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ClearDB()
	},
}

func init() {
	rootCmd.AddCommand(implodeCmd)
}
