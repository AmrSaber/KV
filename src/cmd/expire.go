package cmd

import (
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var expireFlags = struct {
	after time.Duration
	never bool
}{}

var expireCmd = &cobra.Command{
	Use:     "expire <key>",
	Aliases: []string{"ex", "exp"},
	Short:   "Set key expiration",
	Long: `
	Sets expiration for the given key.
	You must use --after to specify expiration time, or use --never to remove any expiration from the key.

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

		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		item := services.GetItem(key)
		if item == nil {
			common.Fail("Key %q does not exist", key)
		}

		if expireFlags.never {
			services.SetValue(key, item.Value, nil, item.IsLocked)
		} else {
			expiresAt := time.Now().Add(expireFlags.after)
			services.SetValue(key, item.Value, &expiresAt, item.IsLocked)
		}
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)

	expireCmd.Flags().DurationVar(&expireFlags.after, "after", 0, "Expires this value after given duration.")
	expireCmd.Flags().BoolVar(&expireFlags.never, "never", false, "Remove any expiration from the key.")

	expireCmd.MarkFlagsMutuallyExclusive("never", "after")
	expireCmd.MarkFlagsOneRequired("never", "after")
}
