package cmd

import (
	"fmt"
	"os"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var backupFlags = struct {
	Path   string
	Stdout bool
}{}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create DB backup",
	Long: `Creates a backup for the database. See info command for backup location.

Created backup can be later restored using 'kv db restore'.

This overwrites any content found at backup path (configurable using --path flag).
`,
	Example: `# Backup DB
kv db backup

# With path
kv db backup --path /some/backup/file

# Write to stdout
kv db backup --stdout | zip backup.zip

# Restore created backup
kv db restore`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		backupPath := backupFlags.Path

		backupWriter := os.Stdout
		if !backupFlags.Stdout {
			var err error
			backupWriter, err = os.Create(backupPath)
			if err != nil {
				common.Fail("Failed to open/create %q for write: %v", backupPath, err)
			}
		}

		defer func() { _ = backupWriter.Close() }()

		err := common.BackupDB(backupWriter)
		if err != nil {
			common.Fail("Failed to create backup: %v", err)
		}

		if !backupFlags.Stdout {
			fmt.Println("Backup created successfully")
		}
	},
}

func init() {
	dbCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&backupFlags.Path, "path", "p", common.GetDefaultBackupPath(), "Backup path")
	backupCmd.Flags().BoolVar(&backupFlags.Stdout, "stdout", false, "Write backup into stdout")

	backupCmd.MarkFlagsMutuallyExclusive("path", "stdout")
}
