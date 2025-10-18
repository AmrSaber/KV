package cmd

import (
	"database/sql"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var deleteFlags = struct {
	prefix bool
	prune  bool
}{}

var deleteCmd = &cobra.Command{
	Use:     "delete <key|prefix>",
	Aliases: []string{"del"},
	Short:   "Delete key or keys prefix",
	GroupID: "kv",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if deleteFlags.prefix {
			common.RunTx(func(tx *sql.Tx) {
				keys := services.MatchExistingKeysByPrefix(tx, key)
				for _, key := range keys {
					services.SetValue(tx, key, "", nil)

					if deleteFlags.prune {
						services.PruneKey(tx, key)
					}
				}
			})

			return
		}

		common.RunTx(func(tx *sql.Tx) {
			if value, _ := services.GetValue(tx, key); value == nil || *value == "" {
				common.Error("Key %q does not exist", key)
				os.Exit(1)
			}

			services.SetValue(tx, key, "", nil)

			if deleteFlags.prune {
				services.PruneKey(tx, key)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVar(&deleteFlags.prefix, "prefix", false, "Delete all keys matching given prefix")
	deleteCmd.Flags().BoolVar(&deleteFlags.prune, "prune", false, "Also delete key(s) history")
}
