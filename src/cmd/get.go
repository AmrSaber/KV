package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var getFlags = struct{ password string }{}

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
		item := services.GetItem(key)

		if item == nil {
			common.Fail("Key %q does not exist", key)
		}

		if item.IsLocked && getFlags.password == "" {
			common.Fail("Key is locked, please pass the password with --password flag")
		}

		value := item.Value
		if getFlags.password != "" {
			var err error
			value, err = common.Decrypt(value, getFlags.password)
			if err != nil {
				common.Fail("Wrong password")
			}
		}

		common.Stdout.Println(value)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getFlags.password, "password", "", "Password to decrypt value if it's encrypted")
}
