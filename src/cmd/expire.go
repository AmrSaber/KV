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
	Use:     "expire <key|key1 key2...>",
	Aliases: []string{"ex", "exp"},
	Short:   "Set or remove expiration for a key or keys",
	Long: `Set or remove expiration for a key or multiple keys.

Use --after with duration suffixes: s (second), m (minute), h (hour)
Example durations: 1h, 30m, 10s, 2h3m4s

Use --never to remove expiration.
Providing a negative duration expires the key immediately.`,
	Example: `  # Set key to expire in 1 hour
  kv expire session-token --after 1h

  # Set multiple keys to expire in 30 minutes
  kv expire temp-data cache-key session-id --after 30m

  # Remove expiration from multiple keys
  kv expire session-token api-key --never

  # Expire key immediately
  kv expire old-token --after -1s`,
	GroupID: "ttl",
	Args:    cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		services.RunInTransaction(func(tx *sql.Tx) {
			for _, key := range args {
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
