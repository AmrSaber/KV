package cmd

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listFlags = struct {
	noValues bool
	output   string
}{}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List keys, optinally matching given prefix.",
	GroupID: "kv",
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeKeyArg(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var prefix string
		if len(args) > 0 {
			prefix = args[0]
		}

		items := services.ListItems(nil, prefix)

		// Sort items by key
		sort.Slice(items, func(i, j int) bool {
			comp := strings.Compare(items[i].Key, items[j].Key)
			return comp < 0
		})

		if listFlags.noValues {
			for i := range items {
				items[i].Value = ""
			}
		}

		switch listFlags.output {
		case "yaml":
			output, _ := yaml.Marshal(items)
			common.Stdout.Println(string(output))
		case "json":
			output, _ := json.MarshalIndent(items, "", "  ")
			common.Stdout.Println(string(output))
		case "table":
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)

			header := []any{"Key", "Value", "Timestamp", "Expires At"}
			if listFlags.noValues {
				header = append(header[0:1], header[2:]...)
			}

			t.AppendHeader(header)

			for _, item := range items {
				expiresAt := "-"
				if item.ExpiresAt != nil {
					expiresAt = item.ExpiresAt.Local().Format(time.DateTime)
				}

				row := []any{
					color.New(color.FgBlue).Sprint(item.Key),
					item.Value,
					color.New(color.FgGreen).Sprint(item.Timestamp.Local().Format(time.DateTime)),
					color.New(color.FgGreen).Sprint(expiresAt),
				}

				if listFlags.noValues {
					row = append(row[0:1], row[2:]...)
				}

				t.AppendRow(row)
			}

			t.SetStyle(table.StyleLight)
			t.Render()
		default:
			common.Error("Unsupported format %q", listFlags.output)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listFlags.noValues, "no-values", "v", false, "Hide values")
	listCmd.Flags().StringVarP(&listFlags.output, "output", "o", "table", "Print format, options: json, yaml, table")
	listCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml", "table"}, cobra.ShellCompDirectiveDefault),
	)
}
