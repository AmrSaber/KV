package cmd

import (
	"os"
	"path"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var dbSetDirectoryCmd = &cobra.Command{
	Use:   "directory <db> <path>",
	Short: "Set database directory",
	Long:  `Move a database to a new directory, backing up the old data first.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		db := args[0]
		directory := args[1]

		common.Assert(db != common.DefaultDBName, "Cannot change directory for default DB")

		directory = common.NormalizePath(directory)
		err := os.MkdirAll(directory, 0o755)
		common.FailOn(err)

		common.BackupDBInPlace(db)

		config := common.GetConfig()
		oldPath := config.GetDBPath(db)
		newPath := path.Join(directory, db+".db")

		err = common.CopyFile(oldPath, newPath)
		if err != nil {
			common.Fail("Failed to move database file: %v", err)
		}

		config.SetDBDirectory(db, directory)

		_ = os.Remove(oldPath)
		_ = os.Remove(oldPath + "-wal")
		_ = os.Remove(oldPath + "-shm")

		common.Stdout.Println("Database directory updated successfully")
	},
}

func init() {
	dbSetCmd.AddCommand(dbSetDirectoryCmd)
}
