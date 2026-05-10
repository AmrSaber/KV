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
	Short: "Copy a key's value to another key and potentially another DB",
	Long: `Copy the value from one key to another key.

The copy operation copies the current value, encryption status, and hidden state from the source key.
TTL is not copied — the destination key will have no expiration unless you set it separately.
If the destination key already exists, it will be updated (creating a new history entry).

Specify a DB on either key using key@db syntax.
As syntactic sugar, <to-key> can take the form '@db-name' which will preserve the same key name.

Copying across DBs loses transactional guarantees — There might be race conditions with other processes
operating on the same keys at the same time.`,
	Example: `  # Copy a key
  kv copy api-key api-key-backup

  # Copy across DBs
  kv copy key@db1 key@db2

  # Copy while preserving the same key name
  kv copy key@db1 @db2

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
		fromKey, fromDB := common.ParseKey(args[0])

		if args[1][0] == '@' {
			args[1] = fromKey + args[1]
		}
		toKey, toDB := common.ParseKey(args[1])

		if fromDB == toDB {
			services.RunInTransaction(fromDB, func(tx *sql.Tx) {
				// Get the source item
				fromItem := services.GetItem(tx, fromKey)
				if fromItem == nil {
					common.Fail("Key %q does not exist", fromKey)
					panic("Unreachable") // To suppress compiler warnings
				}

				// Copy to destination (without TTL)
				services.SetValue(tx, toKey, fromItem.Value, nil, fromItem.IsLocked)
				if fromItem.IsHidden {
					services.HideKey(tx, toKey)
				}
			})
		} else {
			common.PrintCrossDBWarning()

			// Get the source item
			var fromItem *services.KVItem
			services.RunInTransaction(fromDB, func(tx *sql.Tx) {
				fromItem = services.GetItem(tx, fromKey)
				if fromItem == nil {
					common.Fail("Key %q does not exist", fromKey)
					panic("Unreachable") // To suppress compiler warnings
				}
			})

			// Copy to destination (without TTL)
			services.RunInTransaction(toDB, func(tx *sql.Tx) {
				services.SetValue(tx, toKey, fromItem.Value, nil, fromItem.IsLocked)
				if fromItem.IsHidden {
					services.HideKey(tx, toKey)
				}
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
