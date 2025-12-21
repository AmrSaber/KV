package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var restoreFlags = struct {
	Path  string
	Stdin bool
}{}

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore DB backup",
	Long: `Restore database from backup and replace the current database (if exists).

The backup file must be a valid database file created with the 'backup' command.
This will completely replace the current database with the backup.

WARNING: This operation is destructive.
`,

	Example: `# Restore database from default backup location
kv db restore

# Restore from path
kv db restore --path /path/to/backup

# Restore from stdin - you also get UUOC award :)
cat backup.db | kv db restore --stdin`,

	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		backupPath := restoreFlags.Path

		if restoreFlags.Stdin {
			tempFile, err := os.CreateTemp("", "kv-backup")
			common.FailOn(err)

			// Cleanup
			defer func() {
				_ = tempFile.Close()
				_ = os.Remove(tempFile.Name())
			}()

			// Write stdin into a temp file
			_, err = io.Copy(tempFile, os.Stdin)
			common.FailOn(err)

			err = tempFile.Close()
			common.FailOn(err)

			backupPath = tempFile.Name()
		}

		// Check if backup exists
		if _, err := os.Stat(backupPath); err != nil {
			common.Fail("Could not read %q: %v", backupPath, err)
		}

		// Validate backup is a valid SQLite database
		if err := common.ValidateSqliteFile(backupPath); err != nil {
			common.Fail("Invalid backup file: %v", err)
		}

		// Create temp backup in case DB restoration fails
		tempBackupFile, err := os.CreateTemp("", "kv-temp-backup")
		common.FailOn(err)

		defer func() {
			_ = tempBackupFile.Close()
			_ = os.Remove(tempBackupFile.Name())
		}()

		err = common.BackupDB(tempBackupFile)
		if err != nil {
			common.Fail("Could not backup existing database: %v", err)
		}

		// Close database connection
		common.CloseDB()

		// Remove current database and remove WAL files
		dbPath := common.GetDBPath()
		_ = os.Remove(dbPath)
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")

		// Copy backup into DB file
		err = common.CopyFile(backupPath, dbPath)
		if err != nil {
			// Restore backup
			_ = os.Remove(dbPath)
			_ = os.Rename(tempBackupFile.Name(), dbPath)

			common.Fail("Failed to restore database: %v", err)
		}

		// Reopen database to make sure migrations succeed
		_, err = common.GetDB()
		if err != nil {
			// Restore backup
			_ = os.Remove(dbPath)
			_ = os.Rename(tempBackupFile.Name(), dbPath)

			common.Fail("Failed to restore database: %v", err)
		}

		fmt.Println("Database restored from backup successfully")
	},
}

func init() {
	dbCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&restoreFlags.Path, "path", "p", common.GetDefaultBackupPath(), "Existing backup path")
	restoreCmd.Flags().BoolVar(&restoreFlags.Stdin, "stdin", false, "Read from STDIN")

	restoreCmd.MarkFlagsMutuallyExclusive("path", "stdin")
}
