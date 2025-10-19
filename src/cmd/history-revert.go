package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var historyRevertFlags = struct {
	steps int
}{}

var historyRevertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Revert key to an old value and prints it",
	Long: `
		Reverts a key to an old value and prints it.
		By default this reverts the key 1 step back. Use --steps flag to choose the number of steps (same as index from list history command).

		When reverting, a new entry is added to the history with the latest state keeping all the old history as it is.
		So continiously calling revert on a key will only cycle between current and last state polluting the history.
	`,
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

		historyItem := services.GetHistoryItem(key, historyRevertFlags.steps)
		services.SetValue(key, historyItem.Value, nil)

		common.Stdout.Println(historyItem.Value)
	},
}

func init() {
	historyCmd.AddCommand(historyRevertCmd)

	historyRevertCmd.Flags().IntVarP(&historyRevertFlags.steps, "steps", "n", 1, "The number of steps to go back")
}
