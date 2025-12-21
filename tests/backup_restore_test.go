package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AmrSaber/kv/src/common"
)

func TestBackupCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test data
	RunKVSuccess(t, "set", "key1", "value1")
	RunKVSuccess(t, "set", "key2", "value2", "--password", "pass")
	RunKVSuccess(t, "set", "key3", "value3")
	RunKVSuccess(t, "hide", "key3")

	t.Run("backup to default location", func(t *testing.T) {
		backupPath := common.GetDefaultBackupPath()

		RunKVSuccess(t, "db", "backup")

		// Verify file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file was not created")
		}

		// Verify it is valid sqlite file
		if err := common.ValidateSqliteFile(backupPath); err != nil {
			t.Errorf("Invalid backup file: %v", err)
		}
	})

	t.Run("backup to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		backupPath := filepath.Join(tmpDir, "backup.db")

		RunKVSuccess(t, "db", "backup", "--path", backupPath)

		// Verify file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file was not created")
		}

		// Verify it is valid sqlite file
		if err := common.ValidateSqliteFile(backupPath); err != nil {
			t.Errorf("Invalid backup file: %v", err)
		}
	})

	t.Run("backup fails if directory does not exist", func(t *testing.T) {
		backupPath := "/non/existent/path/backup.db"
		RunKVFailure(t, "db", "backup", "--path", backupPath)
	})

	t.Run("backup to stdout", func(t *testing.T) {
		output := RunKVSuccess(t, "db", "backup", "--stdout")

		// Output should be binary data (non-empty)
		if len(output) == 0 {
			t.Error("Expected binary output to stdout")
		}
	})
}

func TestRestoreCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	seedDB := func() {
		RunKVSuccess(t, "set", "original1", "value1")
		RunKVSuccess(t, "set", "original2", "value2", "--password", "pass")
		RunKVSuccess(t, "set", "original3", "value3", "--expires-after", "1h")
		RunKVSuccess(t, "hide", "original3")
		RunKVSuccess(t, "set", "original4", "value4")
		RunKVSuccess(t, "set", "original4", "value4-updated") // Create history
	}

	assertDatabaseRestored := func() {
		output := RunKVSuccess(t, "list", "original2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("original2 should be locked")
		}

		output = RunKVSuccess(t, "get", "original2", "--password", "pass")
		if output != "value2" {
			t.Errorf("Expected 'value2', got: %s", output)
		}

		output = RunKVSuccess(t, "list", "original3")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("original3 should be hidden")
		}

		output = RunKVSuccess(t, "ttl", "original3")
		if !strings.Contains(output, "expires at") {
			t.Error("original3 should have expiration")
		}

		output = RunKVSuccess(t, "history", "list", "original4")
		if !strings.Contains(output, "value4-updated") {
			t.Error("History should contain updated value")
		}
	}

	t.Run("restore from default path", func(t *testing.T) {
		seedDB()

		// Backup
		RunKVSuccess(t, "db", "backup")
		backupStats, _ := os.Stat(common.GetDefaultBackupPath())

		// Modify database
		RunKVSuccess(t, "set", "new-key", "new-value")
		RunKVSuccess(t, "delete", "original1")

		// Restore
		RunKVSuccess(t, "db", "restore")

		// Assert backup is not modified
		stats, err := os.Stat(common.GetDefaultBackupPath())
		if err != nil {
			t.Errorf("Could not read backup stats:%v", err)
		}

		if backupStats.ModTime() != stats.ModTime() {
			t.Error("Backup file was updated after restore")
		}

		assertDatabaseRestored()

		output := RunKVFailure(t, "get", "new-key")
		if !strings.Contains(output, "does not exist") {
			t.Error("new-key should not exist after restore")
		}
	})

	t.Run("restore from custom path", func(t *testing.T) {
		backupFile, err := os.CreateTemp("", "kv-test-backup")
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = backupFile.Close()
			_ = os.Remove(backupFile.Name())
		}()

		err = backupFile.Close()
		if err != nil {
			t.Fatal(err)
		}

		seedDB()

		// Backup
		RunKVSuccess(t, "db", "backup", "--path", backupFile.Name())
		backupStats, _ := os.Stat(common.GetDefaultBackupPath())

		// Modify database
		RunKVSuccess(t, "set", "new-key", "new-value")
		RunKVSuccess(t, "delete", "original1")

		// Restore
		RunKVSuccess(t, "db", "restore", "--path", backupFile.Name())

		// Assert backup is not modified
		stats, err := os.Stat(common.GetDefaultBackupPath())
		if err != nil {
			t.Errorf("Could not read backup stats:%v", err)
		}

		if backupStats.ModTime() != stats.ModTime() {
			t.Error("Backup file was updated after restore")
		}

		assertDatabaseRestored()

		output := RunKVFailure(t, "get", "new-key")
		if !strings.Contains(output, "does not exist") {
			t.Error("new-key should not exist after restore")
		}
	})

	t.Run("restore from stdin", func(t *testing.T) {
		seedDB()

		// Backup
		RunKVSuccess(t, "db", "backup")
		backupStats, _ := os.Stat(common.GetDefaultBackupPath())

		// Modify database
		RunKVSuccess(t, "set", "new-key", "new-value")
		RunKVSuccess(t, "delete", "original1")

		backupFile, err := os.Open(common.GetDefaultBackupPath())
		if err != nil {
			t.Fatal(err)
		}

		// Restore
		RunKVSuccess(t, "db", "restore")
		restoreCmd := RunKVCommand(t, "db", "restore", "--stdin")
		restoreCmd.Stdin = backupFile

		_, err = restoreCmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		// Assert backup is not modified
		stats, err := os.Stat(common.GetDefaultBackupPath())
		if err != nil {
			t.Errorf("Could not read backup stats:%v", err)
		}

		if backupStats.ModTime() != stats.ModTime() {
			t.Error("Backup file was updated after restore")
		}

		assertDatabaseRestored()

		output := RunKVFailure(t, "get", "new-key")
		if !strings.Contains(output, "does not exist") {
			t.Error("new-key should not exist after restore")
		}
	})

	t.Run("restore fails with non-existent file", func(t *testing.T) {
		RunKVFailure(t, "db", "restore", "--path", "/non/existent/file.db")
	})

	t.Run("restore fails with invalid database file", func(t *testing.T) {
		invalidBackup, err := os.CreateTemp("", "kv-test-invalid-backup")
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = invalidBackup.Close()
			_ = os.Remove(invalidBackup.Name())
		}()

		_, err = invalidBackup.Write([]byte("not a database"))
		if err != nil {
			t.Fatal(err)
		}

		err = invalidBackup.Close()
		if err != nil {
			t.Fatal(err)
		}

		RunKVFailure(t, "db", "restore", "--path", invalidBackup.Name())
	})
}
