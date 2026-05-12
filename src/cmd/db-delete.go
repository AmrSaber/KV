package cmd

import (
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var rmFlags = struct {
	prune bool
}{}

var dbRmCmd = &cobra.Command{
	Use:     "delete <db>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a database",
	Long: `Delete a database, backing up the old data first.
Use --prune to delete without backup, also removing any existing backups.`,

	Args: cobra.ExactArgs(1),

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeNonDefaultDBs(toComplete), cobra.ShellCompDirectiveNoFileComp
	},

	Run: func(cmd *cobra.Command, args []string) {
		db := args[0]

		common.Assert(db != common.DefaultDBName, "Cannot delete default DB")

		config := common.GetConfig()

		if _, ok := config.DBs[db]; !ok {
			common.Fail("%q DB does not exist", db)
		}

		dbPath := config.GetDBPath(db)
		backupPath := common.GetDefaultBackupPath(db)

		if rmFlags.prune {
			_ = os.Remove(dbPath)
			_ = os.Remove(dbPath + "-wal")
			_ = os.Remove(dbPath + "-shm")
			_ = os.Remove(backupPath)
		} else {
			common.BackupDBInPlace(db)
			common.Stdout.Printf("Backup created at %s", common.GetDefaultBackupPath(db))

			_ = os.Remove(dbPath)
			_ = os.Remove(dbPath + "-wal")
			_ = os.Remove(dbPath + "-shm")
		}

		config.DeleteDB(db)

		common.Stdout.Printf("Database %q deleted successfully\n", db)
	},
}

func init() {
	dbCmd.AddCommand(dbRmCmd)

	dbRmCmd.Flags().BoolVar(&rmFlags.prune, "prune", false, "Delete without backup")
}
