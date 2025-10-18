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
	quiet   bool
}{}

var ttlCmd = &cobra.Command{
	Use:   "ttl",
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

		value, expiresAt := services.GetValue(nil, key)
		if expiresAt == nil {
			if value != nil && !ttlFlags.quiet {
				common.Stderr.Printf("Key %q does not exipre\n", key)
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

		common.Stdout.Println(expiresAt.Sub(time.Now()).Truncate(time.Second).String())
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)

	ttlCmd.Flags().BoolVarP(&ttlFlags.date, "date", "d", false, "Return expiration date")
	ttlCmd.Flags().BoolVarP(&ttlFlags.seconds, "seconds", "s", false, "Return remaining time in seconds")
	ttlCmd.Flags().BoolVarP(&ttlFlags.quiet, "quiet", "q", false, "Do not print any output")
	ttlCmd.MarkFlagsMutuallyExclusive("date", "seconds", "quiet")
}
