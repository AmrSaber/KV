package cmd

import (
	"database/sql"
	"os"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var ttlFlags = struct {
	date    bool
	seconds bool
	quiet   bool
}{}

var ttlCmd = &cobra.Command{
	Use:   "ttl <key>",
	Short: "Returns how long before the key",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		var value *string
		var expiresAt *time.Time

		common.RunTx(func(tx *sql.Tx) {
			services.CleanUpDB(tx)
			value, expiresAt = services.GetValue(tx, key)
		})

		if expiresAt == nil {
			if !ttlFlags.quiet {
				if value != nil {
					common.Error("Key %q does not exipre", key)
				} else {
					common.Error("Key %q does not exist", key)
				}
			}

			os.Exit(1)
		}

		if ttlFlags.quiet {
			return
		}

		if ttlFlags.seconds {
			common.Stdout.Println(int(expiresAt.Sub(time.Now()).Seconds()))
			return
		}

		if ttlFlags.date {
			common.Stdout.Println(expiresAt.Format(time.DateTime))
			return
		}

		ttl := expiresAt.Sub(time.Now()).Truncate(time.Second).String()
		common.Stdout.Printf("%s (expires at %s)\n", ttl, expiresAt.Format(time.DateTime))
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)

	ttlCmd.Flags().BoolVarP(&ttlFlags.date, "date", "d", false, "Return expiration date")
	ttlCmd.Flags().BoolVarP(&ttlFlags.seconds, "seconds", "s", false, "Return remaining time in seconds")
	ttlCmd.Flags().BoolVarP(&ttlFlags.quiet, "quiet", "q", false, "Do not print any output")
	ttlCmd.MarkFlagsMutuallyExclusive("date", "seconds", "quiet")
}
