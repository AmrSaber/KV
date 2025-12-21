package cmd

import (
	"database/sql"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var ttlFlags = struct {
	date    bool
	seconds bool
}{}

// ttlCmd represents the ttl command
var ttlCmd = &cobra.Command{
	Use:   "ttl <key>",
	Short: "Check the remaining time before a key expires",
	Long: `Check the remaining time before a key expires.

By default, displays the time remaining with expiration date.
Use --date to get only the expiration timestamp.
Use --seconds to get remaining time in seconds (useful for scripts).`,
	Example: `  # Check time remaining (human-readable with date)
  kv ttl session-token
  # Output: 59m56s (expires at 2025-10-20 22:29:25)

  # Get only the expiration date
  kv ttl session-token --date
  # Output: 2025-10-20 22:29:25

  # Get remaining time in seconds
  kv ttl session-token --seconds
  # Output: 3596`,
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

		var value *string
		var expiresAt *time.Time

		services.RunInTransaction(func(tx *sql.Tx) {
			value, expiresAt = services.GetValue(tx, key)
		})

		if expiresAt == nil {
			if value != nil {
				common.Fail("Key %q does not expire", key)
			} else {
				common.Fail("Key %q does not exist", key)
			}
		}

		if ttlFlags.seconds {
			common.Stdout.Println(int(time.Until(*expiresAt).Seconds()))
			return
		}

		if ttlFlags.date {
			common.Stdout.Println(expiresAt.Local().Format(time.DateTime))
			return
		}

		ttl := time.Until(*expiresAt).Truncate(time.Second).String()
		common.Stdout.Printf("%s (expires at %s)\n", ttl, expiresAt.Local().Format(time.DateTime))
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)

	ttlCmd.Flags().BoolVarP(&ttlFlags.date, "date", "d", false, "Return expiration date")
	ttlCmd.Flags().BoolVarP(&ttlFlags.seconds, "seconds", "s", false, "Return remaining time in seconds")
	ttlCmd.MarkFlagsMutuallyExclusive("date", "seconds")
}
