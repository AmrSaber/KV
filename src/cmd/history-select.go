package cmd

import (
	"fmt"
	"slices"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var historySelectFlags = struct{ noValues bool }{}

// historySelectCmd represents the history select command
var historySelectCmd = &cobra.Command{
	Use:   "select <key>",
	Short: "Interactively select and restore a value from history",
	Long: `Interactively select and restore a value from the key's history using a visual selector.

This provides a user-friendly way to browse and choose from previous values.`,
	Example: `  # Interactively select from history
  kv history select api-key

  # Select from history without showing values
  kv history select api-key --no-values`,
	Args: cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if historyPruneFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		items := services.ListKeyHistory(key)
		slices.Reverse(items)

		items = items[1:]

		rows := make([]string, 0, len(items))
		for _, item := range items {
			value := item.Value
			if historySelectFlags.noValues {
				value = ""
			}

			if item.IsLocked {
				value = color.New(color.FgRed).Sprint("[Locked]")
			}

			row := fmt.Sprintf(
				"[%s] %s",
				color.New(color.FgGreen).Sprint(item.Timestamp.Local().Format(time.DateTime)),
				value,
			)

			rows = append(rows, row)
		}

		prompt := promptui.Select{
			Label: fmt.Sprintf("Select a value for %q", key),
			Items: rows,

			Size:         20,
			HideHelp:     true,
			HideSelected: true,
		}

		selectedIndex, _, err := prompt.Run()
		if err != nil {
			common.Fail("")
		}

		selectedItem := items[selectedIndex]

		services.SetValue(key, selectedItem.Value, nil, selectedItem.IsLocked)
		if !selectedItem.IsLocked {
			common.Stdout.Println(selectedItem)
		}
	},
}

func init() {
	historyCmd.AddCommand(historySelectCmd)

	historySelectCmd.Flags().BoolVarP(&historySelectFlags.noValues, "no-values", "v", false, "Hide values")
}
