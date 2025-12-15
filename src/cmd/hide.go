package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var hideFlags = struct {
	prefix bool
}{}

// hideCmd represents the hide command
var hideCmd = &cobra.Command{
	Use:     "hide <key|prefix>",
	Aliases: []string{"obscure", "redact", "conceal"},
	Short:   "Mark a key or keys as hidden",
	Long: `Mark a key or multiple keys as hidden. Hidden keys show as [Hidden] in list and history-list commands.

Note: This does not encrypt the value. Use 'kv lock' for encryption.
Hidden values are still accessible via 'kv get' and can be shown again with 'kv show'.`,
	Example: `  # Hide a single key
  kv hide api-key

  # Hide all keys with a prefix
  kv hide secrets --prefix
  kv hide secrets -p`,
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
			if hideFlags.prefix {
				common.Fail("Prefix must be provided")
			} else {
				common.Fail("Key must be provided")
			}
		}

		key := args[0]

		if hideFlags.prefix {
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.HideKey(tx, item.Key)
				}
			})

			return
		}

		services.RunInTransaction(func(tx *sql.Tx) {
			services.HideKey(tx, key)
		})
	},
}

func init() {
	rootCmd.AddCommand(hideCmd)

	hideCmd.Flags().BoolVarP(&hideFlags.prefix, "prefix", "p", false, "Hide all keys with given prefix")
}
