package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var getFlags = struct{ password string }{}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve the value for the specified key",
	Long: `Retrieve the value for the specified key.

If the key is encrypted, provide the password using --password flag.`,
	Example: `  # Get a plain value
  kv get api-key

  # Get an encrypted value
  kv get github-token --password "mypass"

  # Use in a shell script
  curl -H "Authorization: Bearer $(kv get api-key)" https://api.example.com`,
	GroupID: "kv",
	Args:    cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		var item *services.KVItem

		services.RunInTransaction(func(tx *sql.Tx) {
			item = services.GetItem(tx, key)
		})

		if item == nil {
			common.Fail("Key %q does not exist", key)
			return // To shut up the compiler
		}

		if item.IsLocked && getFlags.password == "" {
			common.Fail("Key is locked, please pass the password with --password flag")
		}

		value := item.Value
		if getFlags.password != "" {
			var err error
			value, err = common.Decrypt(value, getFlags.password)
			if err != nil {
				common.Fail("Wrong password")
			}
		}

		common.Stdout.Println(value)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&getFlags.password, "password", "p", "", "Password to decrypt value if it's encrypted")
}
