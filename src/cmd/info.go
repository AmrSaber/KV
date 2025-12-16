package cmd

import (
	"encoding/json"
	"path/filepath"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var infoFlags = struct{ output string }{}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Displays kv info",
	Long: `Displays kv info and paths.
Note that "backup path" does not mean there is a backup. There might be, there might not.
It just displays the path where a backup would be if there were one.`,

	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		type Info struct {
			DataDir    string `json:"dataDir" yaml:"data-dir"`
			BackupPath string `json:"backupPath" yaml:"backup-path"`

			Config common.Config `json:"config" yaml:"config"`
		}

		info := Info{
			DataDir:    filepath.Dir(common.GetDBPath()),
			BackupPath: common.GetDefaultBackupPath(),
			Config:     common.ReadConfig(),
		}

		switch infoFlags.output {
		case "yaml":
			output, _ := yaml.Marshal(info)
			common.Stdout.Println(string(output))
		case "json":
			output, _ := json.MarshalIndent(info, "", "  ")
			common.Stdout.Println(string(output))
		default:
			common.Fail("Unsupported format %q", infoFlags.output)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().StringVarP(&infoFlags.output, "output", "o", "yaml", "Print format, options: json, yaml")
	_ = infoCmd.RegisterFlagCompletionFunc(
		"output",
		cobra.FixedCompletions([]string{"json", "yaml"}, cobra.ShellCompDirectiveDefault),
	)
}
