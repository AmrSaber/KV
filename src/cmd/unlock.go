package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var unlockFlags = struct {
	prefix   bool
	all      bool
	password string
}{}

// unlockCmd represents the unlock command
var unlockCmd = &cobra.Command{
	Use:     "unlock <key|prefix|key1 key2...>",
	Aliases: []string{"decrypt"},
	Short:   "Decrypt a key or keys back to plain text",
	Long: `Decrypt a key or multiple keys using the provided password, converting them back to plain text.

Note: This removes the latest record from history and replaces it with a plain-text one.`,
	Example: `  # Unlock a single key
  kv unlock api-key --password "mypass"

  # Unlock multiple keys
  kv unlock api-key db-password secret-token --password "mypass"

  # Unlock all keys with a prefix
  kv unlock secrets --prefix --password "mypass"

  # Unlock all keys in the store
  kv unlock --all --password "mypass"`,
	GroupID: "security",
	Args:    cobra.ArbitraryArgs,

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if unlockFlags.all || unlockFlags.prefix {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		if unlockFlags.all {
			if len(args) > 0 {
				common.Fail("Cannot have arguments with --all")
			}

			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, "", services.MatchExisting)
				for _, item := range items {
					err := services.UnlockKey(tx, item.Key, unlockFlags.password)
					if err != nil {
						common.Fail("Wrong password for key %q", item.Key)
					}
				}
			})

			return
		}

		if unlockFlags.prefix {
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
					err := services.UnlockKey(tx, item.Key, unlockFlags.password)
					if err != nil {
						common.Fail("Wrong password for key %q", item.Key)
					}
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
				err := services.UnlockKey(tx, key, unlockFlags.password)
				if err != nil {
					common.Fail("Wrong password")
				}
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)

	unlockCmd.Flags().StringVarP(&unlockFlags.password, "password", "p", "", "Encryption password")
	err := unlockCmd.MarkFlagRequired("password")
	common.FailOn(err)

	unlockCmd.Flags().BoolVar(&unlockFlags.all, "all", false, "Unlock all keys")
	unlockCmd.Flags().BoolVar(&unlockFlags.prefix, "prefix", false, "Unlock all keys with given prefix")
	unlockCmd.MarkFlagsMutuallyExclusive("all", "prefix")
}
