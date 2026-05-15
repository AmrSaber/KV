package cmd

import (
	"database/sql"
	"encoding/json"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listFlags = struct {
	deleted bool
	values  bool
	show    bool
	all     bool

	output string
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
		config := common.GetConfig()

		var prefix, db string
		if len(args) > 0 {
			prefix, db = common.ParseKey(args[0])
		} else {
			db = config.CurrentDB
		}

		matchType := services.MatchExisting
		if listFlags.deleted {
			matchType = services.MatchDeleted
		}

		dbs := []string{db}
		if listFlags.all {
			dbs = slices.Collect(maps.Keys(config.DBs))
		}

		type ListItem struct {
			services.KVItem `yaml:",inline"`
			DB              string `json:"db,omitempty" yaml:"db,omitempty"`
		}

		items := make([]ListItem, 0)
		for _, db := range dbs {
			services.RunInTransaction(db, func(tx *sql.Tx) {
				dbItems := services.ListItems(tx, prefix, matchType)
				listItems := make([]ListItem, 0, len(dbItems))

				for _, dbItem := range dbItems {
					item := ListItem{KVItem: dbItem}

					if listFlags.all {
						item.DB = db
					}

					// Remove the value if any of:
					// - --value flag is not set
					// - item is locked
					// - item is hidden and --show flag is not set
					if !listFlags.values || item.IsLocked || (item.IsHidden && !listFlags.show) {
						item.Value = ""
					}

					listItems = append(listItems, item)
				}

				items = append(items, listItems...)
			})
		}

		if len(items) == 0 {
			if listFlags.deleted {
				common.Stderr.Println("No deleted items.")
			} else {
				common.Stderr.Printf("No saved items in %q DB. Use `kv set` to add one.", config.CurrentDB)
			}

			return
		}

		// Sort items by key
		sort.Slice(items, func(i, j int) bool {
			comp := strings.Compare(items[i].DB, items[j].DB)
			if comp == 0 {
				return strings.Compare(items[i].Key, items[j].Key) < 0
			}

			return comp < 0
		})

		hasExpires, hasLocked := false, false
		for _, item := range items {
			hasExpires = hasExpires || (item.ExpiresAt != nil)
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
			t.SetOutputMirror(common.Stdout.Writer())

			displayDB := listFlags.all
			displayValues := listFlags.values
			displayLocked := hasLocked && !displayValues

			header := []any{"Key"}

			if displayDB {
				header = append(header, "DB")
			}

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

				row := []any{common.Blue(item.Key)}

				if displayDB {
					row = append(row, common.Green(item.DB))
				}

				if displayValues {
					value := item.Value

					// [Locked] takes precedence over [Hidden]
					if item.IsLocked {
						value = common.Red("[Locked]")
					} else if item.IsHidden && !listFlags.show {
						value = common.Red("[Hidden]")
					}

					row = append(row, value)
				}

				row = append(row, common.Green(item.Timestamp.Local().Format(time.DateTime)))

				if hasExpires {
					row = append(row, common.Green(expiresAt))
				}

				if displayLocked {
					isLocked := "-"
					if item.IsLocked {
						isLocked = common.Yellow("Yes")
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

	listCmd.Flags().BoolVarP(&listFlags.values, "values", "v", false, "Show values")
	listCmd.Flags().BoolVarP(&listFlags.deleted, "deleted", "d", false, "List deleted keys")
	listCmd.Flags().BoolVarP(&listFlags.show, "show", "s", false, "Force-show all values")
	listCmd.Flags().BoolVarP(&listFlags.all, "all", "a", false, "List values from all registered DBs")

	listCmd.Flags().StringVarP(&listFlags.output, "output", "o", "table", "Print format, options: json, yaml, table")
	_ = listCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml", "table"}, cobra.ShellCompDirectiveDefault),
	)
}
