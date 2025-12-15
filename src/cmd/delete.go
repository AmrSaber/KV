package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var deleteFlags = struct {
	prefix bool
	prune  bool
}{}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete <key|prefix|key1 key2...>",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a key or keys matching a prefix",
	Long: `Delete a key or multiple keys matching a prefix.

By default, deletion is soft (keeps history). Use --prune to permanently delete including history.`,
	Example: `  # Delete a single key (soft delete, keeps history)
  kv delete api-key

  # Delete multiple keys
  kv delete api-key temp-data old-token

  # Permanently delete multiple keys and their history
  kv delete old-key1 old-key2 --prune

  # Delete all keys with a prefix
  kv delete temp --prefix

  # Permanently delete all keys with a prefix
  kv delete cache --prefix --prune`,
	GroupID: "kv",
	Args:    cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if deleteFlags.prefix {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if deleteFlags.prefix {
			if len(args) > 1 {
				common.Fail("Cannot use --prefix with multiple keys")
			}

			key := args[0]
			services.RunInTransaction(func(tx *sql.Tx) {
				keys := services.ListKeys(tx, key, services.MatchExisting)

				for _, key := range keys {
					services.SetValue(tx, key, "", nil, false)

					if deleteFlags.prune {
						services.PruneKey(tx, key)
					}
				}
			})

			return
		}

		// Handle multiple keys - fail on first error
		services.RunInTransaction(func(tx *sql.Tx) {
			for _, key := range args {
				value, _ := services.GetValue(tx, key)
				if value == nil || *value == "" {
					common.Fail("Key %q does not exist", key)
				}

				services.SetValue(tx, key, "", nil, false)

				if deleteFlags.prune {
					services.PruneKey(tx, key)
				}
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVar(&deleteFlags.prefix, "prefix", false, "Delete all keys matching given prefix")
	deleteCmd.Flags().BoolVar(&deleteFlags.prune, "prune", false, "Also delete key(s) history")
}
