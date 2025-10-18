// Package cmd contains all the commands used.
package cmd

import (
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var rootFlags = struct {
	quiet bool
}{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Your key-value personal store for the CLI",
	Long:  `KV allows you to set and get values in your CLI, with features like history keeping and auto-completion`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		common.Quiet(rootFlags.quiet)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Do not print any output")
}
