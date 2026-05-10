package cmd

import (
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var dbSetNameCmd = &cobra.Command{
	Use:   "name <old> <new>",
	Short: "Rename a database",
	Long:  `Rename a database, backing up the old data first.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]

		common.Assert(oldName != common.DefaultDBName, "Cannot rename default DB")
		common.ValidateDBName(newName)

		config := common.GetConfig()

		if _, ok := config.DBs[oldName]; !ok {
			common.Fail("DB %q does not exist", oldName)
		}

		if _, ok := config.DBs[newName]; ok {
			common.Fail("DB %q already exists", newName)
		}

		common.BackupDBInPlace(oldName)

		err := os.Rename(config.GetDBPath(oldName), config.GetDBPath(newName))
		if err != nil {
			common.Fail("Failed to rename database file: %v", err)
		}

		config.RenameDB(oldName, newName)

		common.Stdout.Println("Database renamed successfully")
	},
}

func init() {
	dbSetCmd.AddCommand(dbSetNameCmd)
}
