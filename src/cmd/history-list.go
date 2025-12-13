package cmd

import (
	"database/sql"
	"encoding/json"
	"os"
	"slices"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var historyListFlags = struct {
	noValues bool
	output   string
	reverse  bool
}{}

// historyListCmd represents the history list command
var historyListCmd = &cobra.Command{
	Use:     "list <key>",
	Aliases: []string{"ls"},
	Short:   "View the complete history for a key",
	Long: `View the complete history for a key, showing all previous values and timestamps.

Index 0 (displayed as "-") indicates the current/latest value.
Higher indices represent older values.`,
	Example: `  # View history for a key
  kv history list api-key

  # View history with JSON output
  kv history list api-key --output json

  # View history in reverse order (oldest first)
  kv history list api-key --reverse

  # View history without values
  kv history list api-key --no-values`,
	Args: cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if historyPruneFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		var kvItems []services.KVItem

		services.RunInTransaction(func(tx *sql.Tx) {
			kvItems = services.ListKeyHistory(tx, key)
		})

		if len(kvItems) == 0 {
			common.Fail("Key %q does not exist", key)
		}

		type IndexedItem struct {
			Index           int `json:"index,omitempty" yaml:"index,omitempty"`
			services.KVItem `json:",inline" yaml:",inline"`
		}

		historyItems := make([]IndexedItem, 0, len(kvItems))
		for i, kvItem := range kvItems {
			item := IndexedItem{
				Index:  len(kvItems) - i - 1,
				KVItem: kvItem,
			}

			// Hide key since they're all for the same key
			item.Key = ""

			if item.IsLocked {
				item.Value = ""
			}

			historyItems = append(historyItems, item)
		}

		if historyListFlags.noValues {
			for i := range historyItems {
				historyItems[i].Value = ""
			}
		}

		hasExpires := false
		for _, item := range historyItems {
			hasExpires = hasExpires || (item.ExpiresAt != nil)
		}

		hasLocked := false
		for _, item := range historyItems {
			hasLocked = hasLocked || item.IsLocked
		}

		if historyListFlags.reverse {
			slices.Reverse(historyItems)
		}

		switch historyListFlags.output {
		case "yaml":
			output, _ := yaml.Marshal(historyItems)
			common.Stdout.Println(string(output))
		case "json":
			output, _ := json.MarshalIndent(historyItems, "", "  ")
			common.Stdout.Println(string(output))
		case "table":
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)

			displayValues := !historyListFlags.noValues
			displayLocked := hasLocked && historyListFlags.noValues

			header := []any{"Index"}

			if displayValues {
				header = append(header, "Value")
			}

			header = append(header, "Timestamp")

			if hasExpires {
				header = append(header, "Expires At")
			}

			if displayLocked {
				header = append(header, "Locked")
			}

			t.AppendHeader(header)

			for _, item := range historyItems {
				expiresAt := "-"
				if item.ExpiresAt != nil {
					expiresAt = item.ExpiresAt.Local().Format(time.DateTime)
				}

				index := color.New(color.FgBlue).Sprint(item.Index)
				if item.Index == 0 {
					index = color.New(color.FgBlue).Sprint("-")
				}

				row := []any{index}

				if displayValues {
					value := item.Value
					if item.IsLocked {
						value = color.New(color.FgRed).Sprint("[Locked]")
					}

					row = append(row, value)
				}

				row = append(row, color.New(color.FgGreen).Sprint(item.Timestamp.Local().Format(time.DateTime)))

				if hasExpires {
					row = append(row, color.New(color.FgGreen).Sprint(expiresAt))
				}

				if displayLocked {
					isLocked := "-"
					if item.IsLocked {
						isLocked = color.New(color.FgYellow).Sprint("Yes")
					}

					row = append(row, isLocked)
				}

				t.AppendRow(row)
			}

			t.SetStyle(table.StyleLight)
			t.Render()
		default:
			common.Fail("Unsupported format %q", historyListFlags.output)
		}
	},
}

func init() {
	historyCmd.AddCommand(historyListCmd)

	historyListCmd.Flags().BoolVarP(&historyListFlags.noValues, "no-values", "v", false, "Hide values")
	historyListCmd.Flags().BoolVarP(&historyListFlags.reverse, "reverse", "r", false, "Reverse history order")
	historyListCmd.Flags().StringVarP(&historyListFlags.output, "output", "o", "table", "Print format, options: json, yaml, table")
	historyListCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml", "table"}, cobra.ShellCompDirectiveDefault),
	)
}
