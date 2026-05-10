package cmd

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/AmrSaber/kv/src/common"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listDBFlags = struct {
	output string
}{}

// restoreCmd represents the restore command
var listDBCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registered DBs",
	Long: `Lists all registered DBs.
Output formats available: table (default), json, yaml`,

	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		type DB struct {
			Name      string `json:"name" yaml:"name"`
			Directory string `json:"directory" yaml:"directory"`
		}
		var items []DB

		for db, dbConfig := range common.GetConfig().DBs {
			items = append(items, DB{Name: db, Directory: dbConfig.Directory})
		}

		// Sort DBs by name ascending
		slices.SortFunc(items, func(a, b DB) int { return strings.Compare(a.Name, b.Name) })

		switch listDBFlags.output {
		case "yaml":
			output, _ := yaml.Marshal(items)
			common.Stdout.Println(string(output))
		case "json":
			output, _ := json.MarshalIndent(items, "", "  ")
			common.Stdout.Println(string(output))
		case "table":
			t := table.NewWriter()
			t.SetStyle(table.StyleLight)
			t.SetOutputMirror(common.Stdout.Writer())
			t.AppendHeader([]any{"DB", "Directory"})

			for _, item := range items {
				t.AppendRow([]any{common.Blue(item.Name), item.Directory})
			}

			t.Render()
		default:
			common.Fail("Unsupported format %q", listDBFlags.output)
		}
	},
}

func init() {
	dbCmd.AddCommand(listDBCmd)

	listDBCmd.Flags().StringVarP(&listDBFlags.output, "output", "o", "table", "Print format, options: json, yaml, table")
	_ = listDBCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml", "table"}, cobra.ShellCompDirectiveDefault),
	)
}
