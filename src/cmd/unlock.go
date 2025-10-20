package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var unlockFlags = struct {
	prefix   bool
	all      bool
	password string
}{}

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Encrypt key(s)",
	Long: `
		Encrypts keys using given password.

		This also removes the latest record from item history and replaces it with a new one
		so that the plain value is no longer in the history.

		If the plain value exists in other records of item history, you may want to prune this key.
	`,
	GroupID: "encryption",
	Args:    cobra.MaximumNArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if unlockFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		var key string
		if !unlockFlags.all {
			if len(args) == 0 {
				if unlockFlags.prefix {
					common.Fail("Prefix must be provided")
				} else {
					common.Fail("Key must be provided")
				}
			}

			key = args[0]
		} else if len(args) > 0 {
			common.Fail("Cannot have an argument with --all")
		}

		if unlockFlags.all || unlockFlags.prefix {
			items := services.ListItems(key, services.MatchExisting)
			for _, item := range items {
				err := services.UnlockKey(item.Key, unlockFlags.password)
				if err != nil {
					common.Fail("Wrong password for key %q", item.Key)
				}
			}

			return
		}

		err := services.UnlockKey(key, unlockFlags.password)
		if err != nil {
			common.Fail("Wrong password")
		}
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)

	unlockCmd.Flags().StringVarP(&unlockFlags.password, "password", "p", "", "Encryption password")
	unlockCmd.MarkFlagRequired("password")

	unlockCmd.Flags().BoolVar(&unlockFlags.all, "all", false, "Unlock all keys")
	unlockCmd.Flags().BoolVar(&unlockFlags.prefix, "prefix", false, "Unlock all keys with given prefix")
	unlockCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
