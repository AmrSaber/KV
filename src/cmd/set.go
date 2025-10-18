package cmd

import (
	"database/sql"
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
	Short: "Set value",
	Long: `
	Assigned the given value to the given key. If no value is provided reads from stdin.

	You can pass --expires-after flag to set the value to expire after given duration.
	Acceptable durations suffixes are: s (second), m (minute), h (hour).
	Example durations: 1h, 30m, 10s, 2h3m4s

	You can also pass --expires-at flag to set the value to expire at given time.
	Acceptable formats are:
	- RFC3339: 2025-01-01T12:34:56Z03:00, 2025-01-01T12:34:56
	- Date-Time: 2025-01-01 12:34:56
	- Date (time is set to 00:00:00): 2025-01-01
	- Time (date is set to today): 12:34:56
	Providing any time in the past expires the key immediately.
	`,
	Args: cobra.RangeArgs(1, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
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
		} else if cmd.Flags().Changed("expires-at") {
			expiresAt = &setFlags.expiresAt
		}

		if value == "" {
			expiresAt = nil
		}

		common.RunTx(func(tx *sql.Tx) {
			services.SetValue(tx, key, value, expiresAt)
			services.CleanUpDB(tx)
		})
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().DurationVar(
		&setFlags.expiresAfter,
		"expires-after",
		0,
		"Expires this value after given duration.",
	)

	setCmd.Flags().TimeVar(
		&setFlags.expiresAt,
		"expires-at",
		time.Time{}, []string{
			time.RFC3339,
			time.DateTime,
			time.DateOnly,
			time.TimeOnly,
		},
		"Expires this value at given time",
	)

	setCmd.MarkFlagsMutuallyExclusive("expires-after", "expires-at")
}
