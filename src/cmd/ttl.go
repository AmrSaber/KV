package cmd

import (
	"os"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var ttlFlags = struct {
	date    bool
	seconds bool
}{}

var ttlCmd = &cobra.Command{
	Use:     "ttl <key>",
	Short:   "Get how long before the key expires",
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

		value, expiresAt = services.GetValue(nil, key)

		if expiresAt == nil {
			if value != nil {
				common.Error("Key %q does not exipre", key)
			} else {
				common.Error("Key %q does not exist", key)
			}

			os.Exit(1)
		}

		if ttlFlags.seconds {
			common.Stdout.Println(int(expiresAt.Sub(time.Now()).Seconds()))
			return
		}

		if ttlFlags.date {
			common.Stdout.Println(expiresAt.Local().Format(time.DateTime))
			return
		}

		ttl := expiresAt.Sub(time.Now()).Truncate(time.Second).String()
		common.Stdout.Printf("%s (expires at %s)\n", ttl, expiresAt.Local().Format(time.DateTime))
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)

	ttlCmd.Flags().BoolVarP(&ttlFlags.date, "date", "d", false, "Return expiration date")
	ttlCmd.Flags().BoolVarP(&ttlFlags.seconds, "seconds", "s", false, "Return remaining time in seconds")
	ttlCmd.MarkFlagsMutuallyExclusive("date", "seconds")
}
