package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var historyRevertFlags = struct {
	steps int
}{}

// historyRevertCmd represents the history revert command
var historyRevertCmd = &cobra.Command{
	Use:   "revert <key>",
	Short: "Revert a key to a previous value",
	Long: `Revert a key to a previous value from its history.

By default, reverts 1 step back. Use --steps to specify the number of steps (same as index from 'kv history list').

Note: Reverting creates a new history entry with the reverted value, preserving all previous history.
Calling revert multiple times without other changes will toggle between the current and previous values.`,
	Example: `  # Revert to previous value (1 step back)
  kv history revert api-key

  # Revert 3 steps back
  kv history revert api-key --steps 3

  # Revert to a specific index from history list
  kv history revert config --steps 5`,
	Args: cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		if historyRevertFlags.steps < 1 {
			common.Fail("steps must be greater than 0, got %v", historyRevertFlags.steps)
		}

		item := services.GetHistoryItem(key, historyRevertFlags.steps)
		services.SetValue(key, item.Value, nil, item.IsLocked)

		if !item.IsLocked {
			common.Stdout.Println(item.Value)
		}
	},
}

func init() {
	historyCmd.AddCommand(historyRevertCmd)

	historyRevertCmd.Flags().IntVarP(&historyRevertFlags.steps, "steps", "n", 1, "The number of steps to go back")
}
