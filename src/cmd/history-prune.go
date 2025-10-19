package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var historyPruneFlags = struct {
	all    bool
	prefix bool
}{}

var historyPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clear history",
	Long: `
		Clears history for:
			- A singel key provided as argument
			- A list of keys prefixed by given argument and --prefix flag
			- All keys

		If a deleted key is targeted for pruning, it's permanently deleted.
	`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if historyPruneFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var key string
		if !historyPruneFlags.all {
			if len(args) == 0 {
				if historyPruneFlags.prefix {
					common.Fail("Prefix must be provided")
				} else {
					common.Fail("Key must be provided")
				}
			}

			key = args[0]
		} else if len(args) > 0 {
			common.Fail("Cannot have an argument with --all")
		}

		if historyPruneFlags.all || historyPruneFlags.prefix {
			services.ClearAllKeysHistory(key)
			return
		}

		services.ClearKeyHistory(key)
	},
}

func init() {
	historyCmd.AddCommand(historyPruneCmd)

	historyPruneCmd.Flags().BoolVar(&historyPruneFlags.all, "all", false, "Prune all keys")
	historyPruneCmd.Flags().BoolVar(&historyPruneFlags.prefix, "prefix", false, "Prune all keys with given prefix")

	historyPruneCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
