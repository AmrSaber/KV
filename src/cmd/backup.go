package cmd

import (
	"fmt"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create DB backup",
	Long: `Creates a backup for the database.

	Created backup can be later restored using 'kv db restore'.

	Only a single backup is kept, so creating a backup removes existing backups.
`,
	Example: `  # Backup DB
	kv db backup

	# Restore created backup
	kv db restore`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := common.BackupDB(common.GetDefaultBackupPath())
		if err != nil {
			common.Fail("Failed to backup current database: %v", err)
		}

		fmt.Println("Database backup created successfully")
	},
}

func init() {
	dbCmd.AddCommand(backupCmd)
}
