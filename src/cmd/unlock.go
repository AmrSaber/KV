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
	Use:   "unlock <key|prefix>",
	Short: "Decrypt a key or keys back to plain text",
	Long: `Decrypt a key or multiple keys using the provided password, converting them back to plain text.

Note: This removes the latest record from history and replaces it with a plain-text one.`,
	Example: `  # Unlock a single key
  kv unlock api-key --password "mypass"

  # Unlock all keys with a prefix
  kv unlock secrets --prefix --password "mypass"

  # Unlock all keys in the store
  kv unlock --all --password "mypass"`,
	GroupID: "security",
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

		services.RunInTransaction(func(tx *sql.Tx) {
			err := services.UnlockKey(tx, key, unlockFlags.password)
			if err != nil {
				common.Fail("Wrong password")
			}
		})
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
