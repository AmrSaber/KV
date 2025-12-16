package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file-path>",
	Short: "Import database from a file",
	Long: `Import a database from a binary file and replace the current database.

The imported file must be a valid database file created with the 'export' command.
This will completely replace the current database with the imported one.

WARNING: This operation is destructive. The current database will be backed up
to <db-path>.backup before importing. You can restore from this backup using
the 'kv db restore' command if needed.

Use "-" as the file path to read from stdin (useful for piping).`,
	Example: `  # Import database from a file
  kv db import backup.db

	# Import from stdin - you also get UUOC award :)
  cat backup.db | kv db import # Or use "-" as file name

  # Import from absolute path
  kv db import /path/to/backup`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourcePath := "-"
		if len(args) > 0 {
			sourcePath = args[0]
		}

		// Handle stdin
		if sourcePath == "-" {
			// Create temporary file for stdin content
			tmpFile, err := os.CreateTemp("", "kv-import-*")
			common.FailOn(err)
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			// Copy stdin to temp file
			_, err = io.Copy(tmpFile, os.Stdin)
			common.FailOn(err)
			err = tmpFile.Close()
			common.FailOn(err)

			sourcePath = tmpFile.Name()
		} else {
			sourcePath = common.NormalizePath(sourcePath)

			// Check if source file exists
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				common.Fail("File does not exist: %s", sourcePath)
			}
		}

		// Validate source is a valid SQLite database
		if err := common.ValidateSqliteFile(sourcePath); err != nil {
			common.Fail("Invalid database file: %v", err)
		}

		// Get destination path
		destPath := common.GetDBPath()
		backupPath := destPath + ".backup"

		// Vacuum current database to commit all WAL changes to main file
		db := common.GetDB()
		_, err := db.Exec("VACUUM")
		common.FailOn(err)

		// Close database connection
		common.CloseDB()

		// Remove old backup if exists
		_ = os.Remove(backupPath)

		// Backup current database if it exists
		if _, err := os.Stat(destPath); err == nil {
			err = os.Rename(destPath, backupPath)
			if err != nil {
				common.Fail("Failed to backup current database: %v", err)
			}

			fmt.Print("Current database backed up")
		}

		// Remove WAL files (should be empty after VACUUM, but remove just in case)
		_ = os.Remove(destPath + "-wal")
		_ = os.Remove(destPath + "-shm")

		// Copy source to destination
		err = common.CopyFile(sourcePath, destPath)
		if err != nil {
			if _, err := os.Stat(backupPath); err == nil {
				_ = os.Remove(destPath)
				_ = os.Rename(backupPath, destPath)
			}

			common.Fail("Failed to import database: %v", err)
		}

		// Reopen database (migrations will run automatically)
		common.GetDB()

		fmt.Println("Database imported successfully")
	},
}

func init() {
	dbCmd.AddCommand(importCmd)
}
