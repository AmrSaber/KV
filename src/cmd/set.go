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
	password     string
}{}

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Store a value for the specified key",
	Long: `Store a value for the specified key. If no value is provided, reads from stdin.

Optionally set automatic expiration using --expires-after with duration suffixes:
  s (second), m (minute), h (hour)
  Example durations: 1h, 30m, 10s, 2h3m4s

Providing a negative duration expires the key immediately.`,
	Example: `  # Store a simple key-value pair
  kv set api-key "sk-1234567890"

  # Store a value with automatic expiration
  kv set session-token "abc123" --expires-after 1h

  # Store encrypted value with password protection
  kv set github-token "ghp_secret" --password "mypass"

  # Store multi-line value from stdin
  echo "line 1\nline 2" | kv set my-config

  # Store JSON configuration
  kv set app.config '{"port": 8080, "debug": true}'`,
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
			common.Fail("No value provided")
		}

		if setFlags.password != "" {
			var err error
			value, err = common.Encrypt(value, setFlags.password)
			common.FailOn(err)
		}

		services.RunInTransaction(func(tx *sql.Tx) {
			services.SetValue(tx, key, value, expiresAt, setFlags.password != "")
		})
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().DurationVar(&setFlags.expiresAfter, "expires-after", 0, "Expires this value after given duration.")
	setCmd.Flags().StringVarP(&setFlags.password, "password", "p", "", "Password to lock this value")
}
