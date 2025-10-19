package cmd

import (
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

var historyListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List key history",
	Args:    cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if historyPruneFlags.all || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(toComplete, services.MatchAll)
	},

	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		kvItems := services.ListKeyHistory(key)

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

			header := []any{"Index"}

			if displayValues {
				header = append(header, "Value")
			}

			header = append(header, "Timestamp")

			if hasExpires {
				header = append(header, "Expires At")
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
					row = append(row, item.Value)
				}

				row = append(row, color.New(color.FgGreen).Sprint(item.Timestamp.Local().Format(time.DateTime)))

				if hasExpires {
					row = append(row, color.New(color.FgGreen).Sprint(expiresAt))
				}

				t.AppendRow(row)
			}

			t.SetStyle(table.StyleLight)
			t.Render()
		default:
			common.Error("Unsupported format %q", historyListFlags.output)
			os.Exit(1)
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
