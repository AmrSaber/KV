package cmd

import (
	"io"
	"os"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var setFlags = struct {
	expiresAfter time.Duration
	expiresAt    time.Time
}{}

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Set value, optionally with a TTL",
	Long: `
	Assigned the given value to the given key. If no value is provided reads from stdin.

	You can pass --expires-after flag to set the value to expire after given duration.
	Acceptable durations suffixes are: s (second), m (minute), h (hour).
	Example durations: 1h, 30m, 10s, 2h3m4s

	Providing any negative duration expires the key immediately.
	`,
	GroupID: "kv",
	Args:    cobra.RangeArgs(1, 2),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := ""
		if len(args) == 2 {
			value = args[1]
		} else {
			stdin, err := io.ReadAll(os.Stdin)
			common.FailOn(err)

			value = string(stdin)
		}

		var expiresAt *time.Time
		if cmd.Flags().Changed("expires-after") {
			expiresAt = &time.Time{}
			*expiresAt = time.Now().Add(setFlags.expiresAfter)
		}

		if value == "" {
			expiresAt = nil
		}

		services.SetValue(key, value, expiresAt)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().DurationVar(&setFlags.expiresAfter, "expires-after", 0, "Expires this value after given duration.")
}
