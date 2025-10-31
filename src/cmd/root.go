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
	Short: "A lightweight, local key-value store for your terminal",
	Long: `KV is a command-line key-value store with encryption, TTL, and version control.

Store configuration, API keys, temporary data, and more—all in your terminal.
Features include AES-256 encryption, automatic expiration, complete history tracking,
and multiple output formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		common.Quiet(rootFlags.quiet)
	},
}

func getVersion() string {
	version := common.GetVersion()
	if version == "" {
		return "??"
	}
	return version
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Set version after it's been potentially injected in main.go
	rootCmd.Version = getVersion()

	common.StartGlobalTransaction()
	services.CleanUpDB()

	err := rootCmd.Execute()
	if err != nil {
		if common.GlobalTx != nil {
			_ = common.GlobalTx.Rollback()
		}

		os.Exit(1)
	}

	services.CleanUpDB()
	_ = common.GlobalTx.Commit()
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{Title: "Key-Value", ID: "kv"},
		&cobra.Group{Title: "TTL", ID: "ttl"},
		&cobra.Group{Title: "Encryption", ID: "encryption"},
	)

	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Do not print any output")
}
