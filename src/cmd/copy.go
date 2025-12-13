package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy <from-key> <to-key>",
	Short: "Copy a key's value to another key",
	Long: `Copy the value from one key to another key.

The copy operation copies the current value and encryption status from the source key.
TTL is not copied - the destination key will have no expiration unless you set it separately.
If the destination key already exists, it will be updated (creating a new history entry).`,
	Example: `  # Copy a key
  kv copy api-key api-key-backup

  # Copy preserves encryption but not TTL
  kv copy encrypted-key encrypted-copy`,
	GroupID: "kv",
	Args:    cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		// Both arguments: complete with existing keys
		if len(args) < 2 {
			return completeKeyArg(toComplete, services.MatchExisting)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		fromKey := args[0]
		toKey := args[1]

		services.RunInTransaction(func(tx *sql.Tx) {
			// Get the source item
			fromItem := services.GetItem(tx, fromKey)
			if fromItem == nil {
				common.Fail("Key %q does not exist", fromKey)
			}

			// Copy to destination (without TTL)
			services.SetValue(tx, toKey, fromItem.Value, nil, fromItem.IsLocked)
		})
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
