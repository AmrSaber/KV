package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var exportFlags = struct{ force bool }{}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export <file-path>",
	Short: "Export database to a file",
	Long: `Export the entire database to a binary file.

The exported file captures the complete state of the database and can be
imported later using the 'import' command to restore the exact state.

Use "-" as the file path to write to stdout (useful for piping).`,
	Example: `  # Export database to a file
  kv db export my-backup

  # Export to stdout
  kv db export > backup.db # Or using "-" as file name

  # Export to absolute path
  kv db export /path/to/backup

  # Overwrite existing file
  kv db export backup.db --force`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Handle stdout
		if len(args) == 0 || args[0] == "-" {
			exportToStdout()
			return
		}

		destPath := common.NormalizePath(args[0])

		// Check if directory exists
		dir := filepath.Dir(destPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			common.Fail("Directory does not exist: %s", dir)
		}

		// Check if file already exists
		if _, err := os.Stat(destPath); err == nil {
			if !exportFlags.force {
				common.Fail("File already exists: %s (use --force to overwrite)", destPath)
			}

			// Remove existing file when --force is used
			err = os.Remove(destPath)
			common.FailOn(err)
		}

		// Export using VACUUM INTO
		db := common.GetDB()
		_, err := db.Exec("VACUUM INTO ?", destPath)
		common.FailOn(err)

		fmt.Printf("Database exported to: %s\n", destPath)
	},
}

func exportToStdout() {
	// Create a temporary file for VACUUM INTO
	tmpFile, err := os.CreateTemp("", "kv-export-*")
	common.FailOn(err)
	tmpPath := tmpFile.Name()
	err = tmpFile.Close()
	common.FailOn(err)
	defer func() { _ = os.Remove(tmpPath) }()

	// Export to temp file
	db := common.GetDB()
	_, err = db.Exec("VACUUM INTO ?", tmpPath)
	common.FailOn(err)

	// Copy temp file to stdout
	file, err := os.Open(tmpPath)
	common.FailOn(err)
	defer func() { _ = file.Close() }()

	_, err = io.Copy(os.Stdout, file)
	common.FailOn(err)
}

func init() {
	dbCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&exportFlags.force, "force", "f", false, "Overwrite existing file if it exists")
}
