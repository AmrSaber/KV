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
	deleted  bool
	noValues bool
	output   string
}{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list [prefix]",
	Aliases: []string{"ls"},
	Short:   "List all keys, optionally matching a prefix",
	Long: `List all keys in the store, optionally filtered by prefix.

Output formats available: table (default), json, yaml
Locked values are displayed as [Locked] in table view.`,
	Example: `  # List all keys
  kv list

  # List keys with a specific prefix
  kv list config

  # List with JSON output
  kv list --output json

  # List keys only (hide values)
  kv list --no-values

  # List deleted keys
  kv list --deleted`,
	GroupID: "kv",
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		if listFlags.deleted {
			return completeKeyArg(toComplete, services.MatchDeleted)
		}

		return completeKeyArg(toComplete, services.MatchExisting)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var prefix string
		if len(args) > 0 {
			prefix = args[0]
		}

		matchType := services.MatchExisting
		if listFlags.deleted {
			matchType = services.MatchDeleted
			listFlags.noValues = true
		}

		items := services.ListItems(prefix, matchType)

		// Sort items by key
		sort.Slice(items, func(i, j int) bool {
			comp := strings.Compare(items[i].Key, items[j].Key)
			return comp < 0
		})

		// Remove the value of locked items
		for i, item := range items {
			if item.IsLocked {
				items[i].Value = ""
			}
		}

		if listFlags.noValues {
			for i := range items {
				items[i].Value = ""
			}
		}

		hasExpires := false
		for _, item := range items {
			hasExpires = hasExpires || (item.ExpiresAt != nil)
		}

		hasLocked := false
		for _, item := range items {
			hasLocked = hasLocked || item.IsLocked
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

			displayValues := !listFlags.noValues
			displayLocked := hasLocked && listFlags.noValues

			header := []any{"Key"}

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

			for _, item := range items {
				expiresAt := "-"
				if item.ExpiresAt != nil {
					expiresAt = item.ExpiresAt.Local().Format(time.DateTime)
				}

				row := []any{color.New(color.FgBlue).Sprint(item.Key)}

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
			common.Fail("Unsupported format %q", listFlags.output)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listFlags.noValues, "no-values", "v", false, "Hide values")
	listCmd.Flags().BoolVarP(&listFlags.deleted, "deleted", "d", false, "List deleted keys")
	listCmd.Flags().StringVarP(&listFlags.output, "output", "o", "table", "Print format, options: json, yaml, table")
	listCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml", "table"}, cobra.ShellCompDirectiveDefault),
	)
}
