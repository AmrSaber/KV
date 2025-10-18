package cmd

import (
	"database/sql"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get stored value",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		var value *string
		common.RunTx(func(tx *sql.Tx) {
			services.CleanUpDB(tx)
			value, _ = services.GetValue(nil, key)
		})

		if value != nil && *value != "" {
			common.Stdout.Println(*value)
		} else {
			common.Error("Key %q does not exist", key)

			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
