package cmd

import (
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
	Aliases: []string{"del", "rm"},
	Short:   "Delete key or keys prefix",
	GroupID: "kv",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if deleteFlags.prefix {
			keys := services.ListKeys(key, services.MatchExisting)
			for _, key := range keys {
				services.SetValue(key, "", nil, false)

				if deleteFlags.prune {
					services.PruneKey(key)
				}
			}

			return
		}

		if value, _ := services.GetValue(key); value == nil || *value == "" {
			common.Fail("Key %q does not exist", key)
		}

		services.SetValue(key, "", nil, false)

		if deleteFlags.prune {
			services.PruneKey(key)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVar(&deleteFlags.prefix, "prefix", false, "Delete all keys matching given prefix")
	deleteCmd.Flags().BoolVar(&deleteFlags.prune, "prune", false, "Also delete key(s) history")
}
