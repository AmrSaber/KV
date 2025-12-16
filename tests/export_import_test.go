package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AmrSaber/kv/src/common"
)

func TestExportCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test data
	RunKVSuccess(t, "set", "key1", "value1")
	RunKVSuccess(t, "set", "key2", "value2", "--password", "pass")
	RunKVSuccess(t, "set", "key3", "value3")
	RunKVSuccess(t, "hide", "key3")

	t.Run("export to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		exportPath := filepath.Join(tmpDir, "backup.db")

		output := RunKVSuccess(t, "db", "export", exportPath)
		if !strings.Contains(output, exportPath) {
			t.Errorf("Expected output to mention export path, got: %s", output)
		}

		// Verify file exists
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Error("Export file was not created")
		}

		// Verify it is valid sqlite file
		if err := common.ValidateSqliteFile(exportPath); err != nil {
			t.Errorf("Invalid exported file: %v", err)
		}
	})

	t.Run("export fails if directory does not exist", func(t *testing.T) {
		exportPath := "/non/existent/path/backup.db"

		output := RunKVFailure(t, "db", "export", exportPath)
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("export fails if file already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		exportPath := filepath.Join(tmpDir, "existing.db")

		// Create file first
		err := os.WriteFile(exportPath, []byte("existing"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		output := RunKVFailure(t, "db", "export", exportPath)
		if !strings.Contains(output, "already exists") {
			t.Errorf("Expected 'already exists' error, got: %s", output)
		}
	})

	t.Run("export with --force overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		exportPath := filepath.Join(tmpDir, "existing.db")

		// Create file first
		err := os.WriteFile(exportPath, []byte("existing"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		RunKVSuccess(t, "db", "export", exportPath, "--force")

		// Verify it is valid sqlite file
		if err := common.ValidateSqliteFile(exportPath); err != nil {
			t.Errorf("Invalid exported file: %v", err)
		}
	})

	t.Run("export to stdout", func(t *testing.T) {
		output := RunKVSuccess(t, "db", "export", "-")

		// Output should be binary data (non-empty)
		if len(output) == 0 {
			t.Error("Expected binary output to stdout")
		}
	})
}

func TestImportCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test data and export
	RunKVSuccess(t, "set", "original1", "value1")
	RunKVSuccess(t, "set", "original2", "value2", "--password", "pass")
	RunKVSuccess(t, "set", "original3", "value3", "--expires-after", "1h")
	RunKVSuccess(t, "hide", "original3")
	RunKVSuccess(t, "set", "original4", "value4")
	RunKVSuccess(t, "set", "original4", "value4-updated") // Create history

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "backup.db")
	RunKVSuccess(t, "db", "export", exportPath)

	t.Run("import from file restores database", func(t *testing.T) {
		// Modify database
		RunKVSuccess(t, "set", "new-key", "new-value")
		RunKVSuccess(t, "delete", "original1")

		// Verify changes
		output := RunKVSuccess(t, "get", "new-key")
		if output != "new-value" {
			t.Error("new-key should exist before import")
		}

		// Import
		output = RunKVSuccess(t, "db", "import", exportPath)
		if !strings.Contains(output, "imported successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify restoration
		output = RunKVSuccess(t, "get", "original1")
		if output != "value1" {
			t.Errorf("Expected 'value1', got: %s", output)
		}

		output = RunKVFailure(t, "get", "new-key")
		if !strings.Contains(output, "does not exist") {
			t.Error("new-key should not exist after import")
		}
	})

	t.Run("import preserves locked keys", func(t *testing.T) {
		output := RunKVSuccess(t, "list", "original2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("original2 should be locked")
		}

		output = RunKVSuccess(t, "get", "original2", "--password", "pass")
		if output != "value2" {
			t.Errorf("Expected 'value2', got: %s", output)
		}
	})

	t.Run("import preserves hidden keys", func(t *testing.T) {
		output := RunKVSuccess(t, "list", "original3")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("original3 should be hidden")
		}
	})

	t.Run("import preserves TTL", func(t *testing.T) {
		output := RunKVSuccess(t, "ttl", "original3")
		if !strings.Contains(output, "expires at") {
			t.Error("original3 should have expiration")
		}
	})

	t.Run("import preserves history", func(t *testing.T) {
		output := RunKVSuccess(t, "history", "list", "original4")
		if !strings.Contains(output, "value4-updated") {
			t.Error("History should contain updated value")
		}
	})

	t.Run("import creates backup", func(t *testing.T) {
		// Create new export
		exportPath2 := filepath.Join(tmpDir, "backup2.db")
		RunKVSuccess(t, "set", "backup-test", "data")
		RunKVSuccess(t, "db", "export", exportPath2)

		// Import original
		output := RunKVSuccess(t, "db", "import", exportPath)
		if !strings.Contains(output, "backed up") {
			t.Error("Should mention backup creation")
		}
	})

	t.Run("import fails with non-existent file", func(t *testing.T) {
		output := RunKVFailure(t, "db", "import", "/non/existent/file.db")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("import fails with invalid database file", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.db")
		err := os.WriteFile(invalidPath, []byte("not a database"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		output := RunKVFailure(t, "db", "import", invalidPath)
		if !strings.Contains(output, "Invalid database") {
			t.Errorf("Expected 'Invalid database' error, got: %s", output)
		}
	})
}

func TestExportImportRoundTrip(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	// Create comprehensive test data
	RunKVSuccess(t, "set", "plain", "plain-value")
	RunKVSuccess(t, "set", "locked", "locked-value", "--password", "mypass")
	RunKVSuccess(t, "set", "hidden", "hidden-value")
	RunKVSuccess(t, "hide", "hidden")
	RunKVSuccess(t, "set", "expiring", "expiring-value", "--expires-after", "2h")
	RunKVSuccess(t, "set", "versioned", "v1")
	RunKVSuccess(t, "set", "versioned", "v2")
	RunKVSuccess(t, "set", "versioned", "v3")

	// Export
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "roundtrip.db")
	RunKVSuccess(t, "db", "export", exportPath)

	// Destroy database
	RunKVSuccess(t, "delete", "plain")
	RunKVSuccess(t, "delete", "locked")
	RunKVSuccess(t, "delete", "hidden")
	RunKVSuccess(t, "delete", "expiring")
	RunKVSuccess(t, "delete", "versioned", "--prune")

	// Import
	RunKVSuccess(t, "db", "import", exportPath)

	// Verify all data
	output := RunKVSuccess(t, "get", "plain")
	if output != "plain-value" {
		t.Errorf("Expected 'plain-value', got: %s", output)
	}

	output = RunKVSuccess(t, "get", "locked", "--password", "mypass")
	if output != "locked-value" {
		t.Errorf("Expected 'locked-value', got: %s", output)
	}

	output = RunKVSuccess(t, "get", "hidden")
	if output != "hidden-value" {
		t.Errorf("Expected 'hidden-value', got: %s", output)
	}

	output = RunKVSuccess(t, "list", "hidden")
	if !strings.Contains(output, "[Hidden]") {
		t.Error("hidden should be marked as hidden")
	}

	output = RunKVSuccess(t, "get", "expiring")
	if output != "expiring-value" {
		t.Errorf("Expected 'expiring-value', got: %s", output)
	}

	output = RunKVSuccess(t, "ttl", "expiring")
	if !strings.Contains(output, "expires at") {
		t.Error("expiring should have TTL")
	}

	output = RunKVSuccess(t, "get", "versioned")
	if output != "v3" {
		t.Errorf("Expected 'v3', got: %s", output)
	}

	output = RunKVSuccess(t, "history", "list", "versioned")
	if !strings.Contains(output, "v1") || !strings.Contains(output, "v2") || !strings.Contains(output, "v3") {
		t.Error("History should contain all versions")
	}
}

func TestRestoreCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("restore fails when no backup exists", func(t *testing.T) {
		output := RunKVFailure(t, "db", "restore")
		if !strings.Contains(output, "No backup file found") {
			t.Errorf("Expected 'No backup file found' error, got: %s", output)
		}
	})

	// Setup test data and create a backup via import
	RunKVSuccess(t, "set", "original1", "value1")
	RunKVSuccess(t, "set", "original2", "value2", "--password", "pass")
	RunKVSuccess(t, "set", "original3", "value3")
	RunKVSuccess(t, "hide", "original3")

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "backup.db")
	RunKVSuccess(t, "db", "export", exportPath)

	// Import to create a backup
	RunKVSuccess(t, "db", "import", exportPath)

	t.Run("restore fails with invalid backup file", func(t *testing.T) {
		// Get DB path and backup path
		dbPath := common.GetDBPath()
		backupPath := dbPath + ".backup"

		// Save current backup
		validBackupPath := backupPath + ".valid"
		err := common.CopyFile(backupPath, validBackupPath)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			// Restore valid backup
			_ = os.Remove(backupPath)
			_ = common.CopyFile(validBackupPath, backupPath)
			_ = os.Remove(validBackupPath)
		}()

		// Replace backup with invalid content
		err = os.WriteFile(backupPath, []byte("not a database"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		output := RunKVFailure(t, "db", "restore")
		if !strings.Contains(output, "Invalid backup") {
			t.Errorf("Expected 'Invalid backup' error, got: %s", output)
		}
	})

	t.Run("restore successfully restores from backup", func(t *testing.T) {
		// Modify database
		RunKVSuccess(t, "set", "new-key", "new-value")
		RunKVSuccess(t, "delete", "original1")

		// Verify changes
		output := RunKVSuccess(t, "get", "new-key")
		if output != "new-value" {
			t.Error("new-key should exist before restore")
		}

		RunKVFailure(t, "get", "original1")

		// Restore
		output = RunKVSuccess(t, "db", "restore")
		if !strings.Contains(output, "restored from") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify restoration
		output = RunKVSuccess(t, "get", "original1")
		if output != "value1" {
			t.Errorf("Expected 'value1', got: %s", output)
		}

		output = RunKVFailure(t, "get", "new-key")
		if !strings.Contains(output, "does not exist") {
			t.Error("new-key should not exist after restore")
		}
	})

	t.Run("restore preserves backup file", func(t *testing.T) {
		dbPath := common.GetDBPath()
		backupPath := dbPath + ".backup"

		// Verify backup still exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file should still exist after restore")
		}

		// Verify it's still valid
		if err := common.ValidateSqliteFile(backupPath); err != nil {
			t.Errorf("Backup file should still be valid: %v", err)
		}
	})

	t.Run("restore preserves all data attributes", func(t *testing.T) {
		// Verify locked key
		output := RunKVSuccess(t, "list", "original2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("original2 should be locked")
		}

		output = RunKVSuccess(t, "get", "original2", "--password", "pass")
		if output != "value2" {
			t.Errorf("Expected 'value2', got: %s", output)
		}

		// Verify hidden key
		output = RunKVSuccess(t, "list", "original3")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("original3 should be hidden")
		}
	})
}

func TestBackupCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	// Create comprehensive test data
	RunKVSuccess(t, "set", "plain", "plain-value")
	RunKVSuccess(t, "set", "locked", "locked-value", "--password", "mypass")
	RunKVSuccess(t, "set", "hidden", "hidden-value")
	RunKVSuccess(t, "hide", "hidden")
	RunKVSuccess(t, "set", "expiring", "expiring-value", "--expires-after", "2h")
	RunKVSuccess(t, "set", "versioned", "v1")
	RunKVSuccess(t, "set", "versioned", "v2")
	RunKVSuccess(t, "set", "versioned", "v3")

	// Backup
	RunKVSuccess(t, "db", "backup")

	// Destroy data
	RunKVSuccess(t, "delete", "plain")
	RunKVSuccess(t, "delete", "locked")
	RunKVSuccess(t, "delete", "hidden")
	RunKVSuccess(t, "delete", "expiring")
	RunKVSuccess(t, "delete", "versioned", "--prune")

	// Restore
	RunKVSuccess(t, "db", "restore")

	// Verify all data
	output := RunKVSuccess(t, "get", "plain")
	if output != "plain-value" {
		t.Errorf("Expected 'plain-value', got: %s", output)
	}

	output = RunKVSuccess(t, "get", "locked", "--password", "mypass")
	if output != "locked-value" {
		t.Errorf("Expected 'locked-value', got: %s", output)
	}

	output = RunKVSuccess(t, "get", "hidden")
	if output != "hidden-value" {
		t.Errorf("Expected 'hidden-value', got: %s", output)
	}

	output = RunKVSuccess(t, "list", "hidden")
	if !strings.Contains(output, "[Hidden]") {
		t.Error("hidden should be marked as hidden")
	}

	output = RunKVSuccess(t, "get", "expiring")
	if output != "expiring-value" {
		t.Errorf("Expected 'expiring-value', got: %s", output)
	}

	output = RunKVSuccess(t, "ttl", "expiring")
	if !strings.Contains(output, "expires at") {
		t.Error("expiring should have TTL")
	}

	output = RunKVSuccess(t, "get", "versioned")
	if output != "v3" {
		t.Errorf("Expected 'v3', got: %s", output)
	}

	output = RunKVSuccess(t, "history", "list", "versioned")
	if !strings.Contains(output, "v1") || !strings.Contains(output, "v2") || !strings.Contains(output, "v3") {
		t.Error("History should contain all versions")
	}
}
