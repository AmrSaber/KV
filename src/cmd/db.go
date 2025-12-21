package cmd

import "github.com/spf13/cobra"

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operations",
	Long: `
Operations related to kv's database.

Note: These commands are not thread-safe. It's the responsibility of the caller to make sure no other commands run at the same time.
	`,
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
