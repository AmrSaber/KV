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
	Use:     "lock <key|prefix|key1 key2...>",
	Aliases: []string{"encrypt"},
	Short:   "Encrypt a key or keys with password protection",
	Long: `Encrypt a key or multiple keys using AES-256-GCM encryption with the provided password.

Note: This removes the latest record from history and replaces it with an encrypted one.
If plain-text values exist in older history records, consider using 'kv history prune' to remove them.`,
	Example: `  # Lock a single key
  kv lock api-key --password "mypass"

  # Lock multiple keys
  kv lock api-key db-password secret-token --password "mypass"

  # Lock all keys with a prefix
  kv lock secrets --prefix --password "mypass"

  # Lock all keys in the store
  kv lock --all --password "mypass"`,
	GroupID: "security",
	Args:    cobra.ArbitraryArgs,

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if lockFlags.all || lockFlags.prefix {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		if lockFlags.all {
			if len(args) > 0 {
				common.Fail("Cannot have arguments with --all")
			}

			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, "", services.MatchExisting)
				for _, item := range items {
					services.LockKey(tx, item.Key, lockFlags.password)
				}
			})

			return
		}

		if lockFlags.prefix {
			if len(args) == 0 {
				common.Fail("Prefix must be provided")
			}
			if len(args) > 1 {
				common.Fail("Cannot use --prefix with multiple keys")
			}

			key := args[0]
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.LockKey(tx, item.Key, lockFlags.password)
				}
			})

			return
		}

		// Handle multiple keys - fail on first error
		if len(args) == 0 {
			common.Fail("At least one key must be provided")
		}

		services.RunInTransaction(func(tx *sql.Tx) {
			for _, key := range args {
				services.LockKey(tx, key, lockFlags.password)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)

	lockCmd.Flags().StringVarP(&lockFlags.password, "password", "p", "", "Encryption password")
	err := lockCmd.MarkFlagRequired("password")
	common.FailOn(err)

	lockCmd.Flags().BoolVar(&lockFlags.all, "all", false, "Lock all keys")
	lockCmd.Flags().BoolVar(&lockFlags.prefix, "prefix", false, "Lock all keys with given prefix")
	lockCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
