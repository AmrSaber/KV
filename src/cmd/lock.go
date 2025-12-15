package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var lockFlags = struct {
	prefix   bool
	all      bool
	password string
}{}

// lockCmd represents the lock command
var lockCmd = &cobra.Command{
	Use:   "lock <key|prefix>",
	Short: "Encrypt a key or keys with password protection",
	Long: `Encrypt a key or multiple keys using AES-256-GCM encryption with the provided password.

Note: This removes the latest record from history and replaces it with an encrypted one.
If plain-text values exist in older history records, consider using 'kv history prune' to remove them.`,
	Example: `  # Lock a single key
  kv lock api-key --password "mypass"

  # Lock all keys with a prefix
  kv lock secrets --prefix --password "mypass"

  # Lock all keys in the store
  kv lock --all --password "mypass"`,
	GroupID: "security",
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
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.LockKey(tx, item.Key, lockFlags.password)
				}
			})

			return
		}

		services.RunInTransaction(func(tx *sql.Tx) {
			services.LockKey(tx, key, lockFlags.password)
		})
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)

	lockCmd.Flags().StringVarP(&lockFlags.password, "password", "p", "", "Encryption password")
	lockCmd.MarkFlagRequired("password")

	lockCmd.Flags().BoolVar(&lockFlags.all, "all", false, "Lock all keys")
	lockCmd.Flags().BoolVar(&lockFlags.prefix, "prefix", false, "Lock all keys with given prefix")
	lockCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
