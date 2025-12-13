package cmd

import (
	"database/sql"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var expireFlags = struct {
	after time.Duration
	never bool
}{}

// expireCmd represents the expire command
var expireCmd = &cobra.Command{
	Use:     "expire <key>",
	Aliases: []string{"ex", "exp"},
	Short:   "Set or remove expiration for a key",
	Long: `Set or remove expiration for a key.

Use --after with duration suffixes: s (second), m (minute), h (hour)
Example durations: 1h, 30m, 10s, 2h3m4s

Use --never to remove expiration.
Providing a negative duration expires the key immediately.`,
	Example: `  # Set key to expire in 1 hour
  kv expire session-token --after 1h

  # Set key to expire in 30 minutes
  kv expire temp-data --after 30m

  # Remove expiration from a key
  kv expire session-token --never

  # Expire key immediately
  kv expire old-token --after -1s`,
	GroupID: "ttl",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		services.RunInTransaction(func(tx *sql.Tx) {
			item := services.GetItem(tx, key)
			if item == nil {
				common.Fail("Key %q does not exist", key)
			}

			if expireFlags.never {
				services.SetValue(tx, key, item.Value, nil, item.IsLocked)
			} else {
				expiresAt := time.Now().Add(expireFlags.after)
				services.SetValue(tx, key, item.Value, &expiresAt, item.IsLocked)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)

	expireCmd.Flags().DurationVar(&expireFlags.after, "after", 0, "Expires this value after given duration.")
	expireCmd.Flags().BoolVar(&expireFlags.never, "never", false, "Remove any expiration from the key.")

	expireCmd.MarkFlagsMutuallyExclusive("never", "after")
	expireCmd.MarkFlagsOneRequired("never", "after")
}
