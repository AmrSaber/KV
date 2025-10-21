package cmd

import (
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename <old-key> <new-key>",
	Short: "Rename a key to a new name",
	Long: `Rename a key by changing its name in the store across all history items.

The rename operation preserves all history, encryption status, TTL, and other metadata.
The old key name will no longer exist after the rename.`,
	Example: `  # Rename a key
  kv rename old-api-key new-api-key

  # Rename preserves all properties including encryption
  kv rename encrypted-secret new-secret-name`,
	GroupID: "kv",
	Args:    cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		// First argument: complete with existing keys
		if len(args) == 0 {
			return completeKeyArg(toComplete, services.MatchExisting)
		}

		// Second argument: no completion
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		oldKey := args[0]
		newKey := args[1]

		services.RenameKey(oldKey, newKey)
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
