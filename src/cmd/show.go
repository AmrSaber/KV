package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var showFlags = struct {
	prefix bool
}{}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show <key|prefix|key1 key2...>",
	Aliases: []string{"stet", "reveal"},
	Short:   "Remove hidden status from a key or keys",
	Long:    `Remove hidden status from a key or multiple keys, making values visible in list and history-list commands.`,
	Example: `  # Show a single key
  kv show api-key

  # Show multiple keys
  kv show api-key db-password secret-token

  # Show all keys with a prefix
  kv show secrets --prefix
  kv show secrets -p`,
	GroupID: "security",
	Args:    cobra.MinimumNArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if showFlags.prefix {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if showFlags.prefix {
				common.Fail("Prefix must be provided")
			} else {
				common.Fail("At least one key must be provided")
			}
		}

		if showFlags.prefix {
			if len(args) > 1 {
				common.Fail("Cannot use --prefix with multiple keys")
			}

			key := args[0]
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.ShowKey(tx, item.Key)
				}
			})

			return
		}

		// Handle multiple keys - fail on first error
		services.RunInTransaction(func(tx *sql.Tx) {
			for _, key := range args {
				services.ShowKey(tx, key)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&showFlags.prefix, "prefix", "p", false, "Show all keys with given prefix")
}
