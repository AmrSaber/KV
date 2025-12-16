package cmd

import "github.com/spf13/cobra"

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operations",
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
