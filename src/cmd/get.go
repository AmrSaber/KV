package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get <key>",
	Short:   "Get stored value",
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
		value, _ := services.GetValue(key)

		if value != nil && *value != "" {
			common.Stdout.Println(*value)
		} else {
			common.Fail("Key %q does not exist", key)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
