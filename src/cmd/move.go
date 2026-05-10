package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
var moveCmd = &cobra.Command{
	Use:     "move <old-key> <new-key>",
	Aliases: []string{"rename", "mv"},
	Short:   "Move a key to a new name and potentially new DB",
	Long: `Move a key to a new name. The key keeps all its history, encryption, TTL, and other metadata.

The old key name will no longer exist after the move.

Specify a DB on either key using key@db syntax.
As syntactic sugar, <new-key> can take the form '@db-name' which will preserve the same key name.

Moving across DBs loses transactional guarantees — if the process crashes mid-operation, the key may exist in both DBs.
Also, there might be race conditions with other processes operating on the same keys at the same time.`,
	Example: `  # Move a key to a new name
  kv move old-api-key new-api-key

  # Move across DBs
  kv move key@db1 key@db2

  # Move while preserving the same key name
  kv move key@db1 @db2`,
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
		oldKey, oldDB := common.ParseKey(args[0])

		if args[1][0] == '@' {
			args[1] = oldKey + args[1]
		}
		newKey, newDB := common.ParseKey(args[1])

		if oldDB == newDB {
			services.RunInTransaction(oldDB, func(tx *sql.Tx) {
				services.MoveKey(tx, oldKey, newKey)
			})
		} else {
			common.PrintCrossDBWarning()

			var items []map[string]any
			services.RunInTransaction(oldDB, func(tx *sql.Tx) {
				items = services.ScanRawKeyRows(tx, oldKey)
			})

			// Rename key and delete IDs
			for i := range items {
				items[i]["key"] = newKey
				delete(items[i], "id")
			}

			services.RunInTransaction(newDB, func(tx *sql.Tx) {
				// Check if key already exists
				newItem := services.GetItem(tx, newKey)
				if newItem != nil {
					common.Fail("Key %q already exists", newKey)
				}

				// Insert new rows
				services.InsertRawRows(tx, items)
			})

			services.RunInTransaction(oldDB, func(tx *sql.Tx) {
				services.PruneKey(tx, oldKey)
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
