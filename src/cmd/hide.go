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
	Use:     "hide <key|prefix|key1 key2...>",
	Aliases: []string{"obscure", "redact", "conceal"},
	Short:   "Mark a key or keys as hidden",
	Long: `Mark a key or multiple keys as hidden. Hidden keys show as [Hidden] in list and history-list commands.

Note: This does not encrypt the value. Use 'kv lock' for encryption.
Hidden values are still accessible via 'kv get' and can be shown again with 'kv show'.`,
	Example: `  # Hide a single key
  kv hide api-key

  # Hide multiple keys
  kv hide api-key db-password secret-token

  # Hide all keys with a prefix
  kv hide secrets --prefix
  kv hide secrets -p`,
	GroupID: "security",
	Args:    cobra.MinimumNArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if hideFlags.prefix {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if hideFlags.prefix {
				common.Fail("Prefix must be provided")
			} else {
				common.Fail("At least one key must be provided")
			}
		}

		if hideFlags.prefix {
			if len(args) > 1 {
				common.Fail("Cannot use --prefix with multiple keys")
			}

			key := args[0]
			services.RunInTransaction(func(tx *sql.Tx) {
				items := services.ListItems(tx, key, services.MatchExisting)
				for _, item := range items {
					services.HideKey(tx, item.Key)
				}
			})

			return
		}

		// Handle multiple keys - fail on first error
		services.RunInTransaction(func(tx *sql.Tx) {
			for _, key := range args {
				services.HideKey(tx, key)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(hideCmd)

	hideCmd.Flags().BoolVarP(&hideFlags.prefix, "prefix", "p", false, "Hide all keys with given prefix")
}
