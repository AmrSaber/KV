package cmd

import "github.com/spf13/cobra"

var dbSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update database settings",
	Long:  `Update database settings such as name, directory, etc.`,
}

func init() {
	dbCmd.AddCommand(dbSetCmd)
}
