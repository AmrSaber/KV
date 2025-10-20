package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var lockFlags = struct {
	prefix   bool
	all      bool
	password string
}{}

var lockCmd = &cobra.Command{
	Use:   "lock",
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
		if lockFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		var key string
		if !lockFlags.all {
			if len(args) == 0 {
				if lockFlags.prefix {
					common.Fail("Prefix must be provided")
				} else {
					common.Fail("Key must be provided")
				}
			}

			key = args[0]
		} else if len(args) > 0 {
			common.Fail("Cannot have an argument with --all")
		}

		if lockFlags.all || lockFlags.prefix {
			items := services.ListItems(key, services.MatchExisting)
			for _, item := range items {
				services.LockKey(item.Key, lockFlags.password)
			}

			return
		}

		services.LockKey(key, lockFlags.password)
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)

	lockCmd.Flags().StringVar(&lockFlags.password, "with", "", "Encryption password")
	lockCmd.MarkFlagRequired("with")

	lockCmd.Flags().BoolVar(&lockFlags.all, "all", false, "Lock all keys")
	lockCmd.Flags().BoolVar(&lockFlags.prefix, "prefix", false, "Lock all keys with given prefix")
	lockCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
