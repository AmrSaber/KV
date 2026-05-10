package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:     "rename <old-key> <new-key>",
	Aliases: []string{"move", "mv"},
	Short:   "Rename a key to a new name",
	Long: `Rename a key by changing its name in the store across all history items.

The rename operation preserves all history, encryption status, TTL, and other metadata.
The old key name will no longer exist after the rename.

Moving keys across DBs will move the whole history to the new DB under the new name in the new DB.
As syntactic sugar, <new-key> can take the form '@db-name' which will preserve the same key name.`,
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
		oldKey, oldDB := common.ParseKey(args[0])

		if args[1][0] == '@' {
			args[1] = oldKey + args[1]
		}
		newKey, newDB := common.ParseKey(args[1])

		if oldDB == newDB {
			services.RunInTransaction(oldDB, func(tx *sql.Tx) {
				services.RenameKey(tx, oldKey, newKey)
			})
		} else {
			common.PrintCrossDBWarning()

			var items []map[string]any
			services.RunInTransaction(oldDB, func(tx *sql.Tx) {
				items = services.ScanRawKeyRows(tx, oldKey)
			})

			// Rename and delete IDs
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
	rootCmd.AddCommand(renameCmd)
}
