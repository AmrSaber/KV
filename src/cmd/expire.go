package cmd

import (
	"database/sql"
	"os"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var expireFlags = struct {
	after time.Duration
	at    time.Time
}{}

var expireCmd = &cobra.Command{
	Use:     "expire <key>",
	Aliases: []string{"ex", "exp"},
	Short:   "Set key expiration",
	Long: `
	Sets expiration for the given key. You must use --after to specify expiration time.

	For --after, acceptable durations suffixes are: s (second), m (minute), h (hour).
	Example durations: 1h, 30m, 10s, 2h3m4s

	Providing any negative duration expires the key immediately.
	`,
	GroupID: "ttl",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		expiresAt := time.Now().Add(expireFlags.after)

		common.RunTx(func(tx *sql.Tx) {
			value, _ := services.GetValue(tx, key)
			if value == nil || *value == "" {
				common.Error("Key %q does not exist", key)
				os.Exit(1)
			}

			services.SetValue(tx, key, *value, &expiresAt)
		})
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)

	expireCmd.Flags().DurationVar(&expireFlags.after, "after", 0, "Expires this value after given duration.")
	expireCmd.MarkFlagRequired("after")
}
