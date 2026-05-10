// Package cmd contains all the commands used.
package cmd

import (
	"maps"
	"os"
	"slices"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var rootFlags = struct {
	quiet bool
	db    string
}{}

var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "A lightweight, local key-value store for your terminal",
	Long: `KV is a command-line key-value store with encryption, TTL, and version control.

Store configuration, API keys, temporary data, and more—all in your terminal.
Features include AES-256 encryption, automatic expiration, complete history tracking,
and multiple output formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		common.Quiet(rootFlags.quiet)

		// --db flag takes precedence over KV_DB env variable and default DB
		dbFlag := cmd.Flag("db")
		if dbFlag.Changed {
			common.GetConfig().CurrentDB = dbFlag.Value.String()
		}
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
	defer func() {
		for name := range common.CachedDBs {
			common.GetConfig().RegisterDB(name)
		}

		common.CloseDBs()

		if err := recover(); err != nil {
			common.Stderr.Println(common.Red(err))
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{Title: "Key-Value", ID: "kv"},
		&cobra.Group{Title: "TTL", ID: "ttl"},
		&cobra.Group{Title: "Security", ID: "security"},
	)

	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Do not print any output")
	rootCmd.PersistentFlags().StringVar(&rootFlags.db, "db", common.DefaultDBName, "Database name")

	_ = rootCmd.RegisterFlagCompletionFunc(
		"db",
		cobra.FixedCompletions(slices.Collect(maps.Keys(common.GetConfig().DBs)), cobra.ShellCompDirectiveDefault),
	)
}
