package cmd

import (
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get stored value",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			toComplete = ""
		}

		matchingKeys := services.GetKeysMatchingPrefix(nil, toComplete)

		return []cobra.Completion(matchingKeys), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value, _ := services.GetValue(nil, key)
		if value != nil && *value != "" {
			common.Stdout.Println(*value)
		} else {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
