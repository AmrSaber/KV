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
	Use:     "show <key|prefix>",
	Aliases: []string{"stet", "reveal"},
	Short:   "Remove hidden status from a key or keys",
	Long:    `Remove hidden status from a key or multiple keys, making values visible in list and history-list commands.`,
	Example: `  # Show a single key
  kv show api-key

  # Show all keys with a prefix
  kv show secrets --prefix
  kv show secrets -p`,
	GroupID: "security",
	Args:    cobra.MaximumNArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if showFlags.prefix {
				common.Fail("Prefix must be provided")
			} else {
				common.Fail("Key must be provided")
			}
		}

		key := args[0]

		if showFlags.prefix {
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.ShowKey(tx, item.Key)
				}
			})

			return
		}

		services.RunInTransaction(func(tx *sql.Tx) {
			services.ShowKey(tx, key)
		})
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&showFlags.prefix, "prefix", "p", false, "Show all keys with given prefix")
}
