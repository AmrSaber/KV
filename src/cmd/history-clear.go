package cmd

import (
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear key history",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		services.ClearKeyHistory(nil, key)
	},
}

func init() {
	historyCmd.AddCommand(historyClearCmd)
}
