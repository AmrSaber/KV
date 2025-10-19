// Package cmd contains all the commands used.
package cmd

import (
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

var rootFlags = struct{ quiet bool }{}

var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Your key-value personal store for the CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		common.Quiet(rootFlags.quiet)
		common.StartGlobalTransaction()

		services.CleanUpDB()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		services.CleanUpDB()

		common.GlobalTx.Commit()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if common.GlobalTx != nil {
			common.GlobalTx.Rollback()
		}

		common.Error("%v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{Title: "Key-Value", ID: "kv"},
		&cobra.Group{Title: "TTL", ID: "ttl"},
	)

	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Do not print any output")
}
