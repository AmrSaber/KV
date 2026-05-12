package tests

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"
)

func TestDBRmCommand(t *testing.T) {
	t.Run("deletes DB with backup", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		dataDir := path.Join(tmpDir, "kv")

		RunKVSuccess(t, "--db", "testdb", "set", "k", "v")

		dbPath := path.Join(dataDir, "testdb.db")
		backupPath := path.Join(dataDir, "testdb.db.backup")

		if _, err := os.Stat(dbPath); err != nil {
			t.Fatalf("testdb.db should exist before deletion: %v", err)
		}
		if _, err := os.Stat(backupPath); err == nil {
			t.Fatal("testdb.db.backup should not exist before deletion")
		}

		RunKVSuccess(t, "db", "delete", "testdb")

		if _, err := os.Stat(dbPath); err == nil {
			t.Error("testdb.db should be deleted")
		}
		if _, err := os.Stat(backupPath); err != nil {
			t.Errorf("testdb.db.backup should exist after deletion: %v", err)
		}

		output := RunKVSuccess(t, "db", "list", "--output", "json")
		var items []map[string]string
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("db list output is not valid JSON: %v", err)
		}
		for _, item := range items {
			if item["name"] == "testdb" {
				t.Error("testdb should not appear in DB list after deletion")
				break
			}
		}
	})

	t.Run("deletes DB with --prune", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		dataDir := path.Join(tmpDir, "kv")

		RunKVSuccess(t, "--db", "testdb", "set", "k", "v")

		dbPath := path.Join(dataDir, "testdb.db")
		if _, err := os.Stat(dbPath); err != nil {
			t.Fatalf("testdb.db should exist before deletion: %v", err)
		}

		RunKVSuccess(t, "db", "delete", "testdb", "--prune")

		if _, err := os.Stat(dbPath); err == nil {
			t.Error("testdb.db should be deleted")
		}
		if _, err := os.Stat(path.Join(dataDir, "testdb.db.backup")); err == nil {
			t.Error("No backup should be created with --prune")
		}
	})

	t.Run("--prune removes existing backup", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		dataDir := path.Join(tmpDir, "kv")

		RunKVSuccess(t, "--db", "testdb", "set", "k", "v")

		backupPath := path.Join(dataDir, "testdb.db.backup")
		dbBytes, err := os.ReadFile(path.Join(dataDir, "testdb.db"))
		if err != nil {
			t.Fatalf("Failed to read testdb.db: %v", err)
		}
		if err := os.WriteFile(backupPath, dbBytes, 0o644); err != nil {
			t.Fatalf("Failed to create testdb.db.backup: %v", err)
		}
		if _, err := os.Stat(backupPath); err != nil {
			t.Fatalf("Backup should exist before prune: %v", err)
		}

		RunKVSuccess(t, "db", "delete", "testdb", "--prune")

		if _, err := os.Stat(path.Join(dataDir, "testdb.db")); err == nil {
			t.Error("testdb.db should be deleted after --prune")
		}
		if _, err := os.Stat(backupPath); err == nil {
			t.Error("Existing backup should also be deleted with --prune")
		}
	})

	t.Run("fails on default DB", func(t *testing.T) {
		SetupTestDB(t)

		output := RunKVFailure(t, "db", "delete", "default")
		if !strings.Contains(output, "Cannot delete default DB") {
			t.Errorf("Expected error about default DB, got: %s", output)
		}
	})

	t.Run("fails on non-existent DB", func(t *testing.T) {
		SetupTestDB(t)

		output := RunKVFailure(t, "db", "delete", "nonexistent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected error about non-existent DB, got: %s", output)
		}
	})
}
