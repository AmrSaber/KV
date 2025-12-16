package cmd

import (
	"fmt"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from automatic backup",
	Long: `Restore the database from the automatic backup file created during import.

When you run 'kv db import', a backup of the current database is automatically
created at <db-path>.backup. This command restores that backup.

The backup file is preserved after restoration, allowing you to restore again
if needed.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get database path
		dbPath := common.GetDBPath()
		backupPath := common.GetDefaultBackupPath()

		// Check if backup exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			common.Fail("No backup file found")
		}

		// Validate backup is a valid SQLite database
		if err := common.ValidateSqliteFile(backupPath); err != nil {
			common.Fail("Invalid backup file: %v", err)
		}

		// Close database connection
		common.CloseDB()

		// Remove current database and remove WAL files
		_ = os.Remove(dbPath)
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")

		// Copy backup to current location
		err := common.CopyFile(backupPath, dbPath)
		if err != nil {
			common.Fail("Failed to restore from backup: %v", err)
		}

		// Reopen database (migrations will run automatically)
		_, err = common.GetDB()
		if err != nil {
			common.Fail("Failed to restore from backup: %v", err)
		}

		fmt.Println("Database restored from backup")
	},
}

func init() {
	dbCmd.AddCommand(restoreCmd)
}
