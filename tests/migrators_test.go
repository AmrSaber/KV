package tests

import (
	"os"
	"path"
	"strings"
	"testing"
)

func runKVWithEnv(t *testing.T, extraEnv []string, args ...string) (string, error) {
	t.Helper()

	cmd := RunKVCommand(t, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func TestMigrateOldDBName(t *testing.T) {
	t.Run("no warning when no kv.db exists", func(t *testing.T) {
		SetupTestDB(t)

		output, err := RunKV(t, "get", "nonexistent")
		_ = err
		if strings.Contains(output, "You have both") {
			t.Error("There should be no migration warning with no kv.db present")
		}
	})

	t.Run("renames kv.db to default and migrates data", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		dataDir := path.Join(tmpDir, "kv")

		_, err := runKVWithEnv(t, []string{"KV_DB=kv"}, "set", "migrated-key", "migrated-value")
		if err != nil {
			t.Fatalf("Failed to seed kv.db: %v", err)
		}

		RunKVSuccess(t, "set", "normal-key", "normal-value")

		if got := RunKVSuccess(t, "get", "migrated-key"); got != "migrated-value" {
			t.Errorf("Expected migrated data 'migrated-value', got: %s", got)
		}
		if got := RunKVSuccess(t, "get", "normal-key"); got != "normal-value" {
			t.Errorf("Expected normal data 'normal-value', got: %s", got)
		}

		if _, err := os.Stat(path.Join(dataDir, "default.db")); err != nil {
			t.Errorf("default.db should exist after migration: %v", err)
		}
		if _, err := os.Stat(path.Join(dataDir, "kv.db")); err == nil {
			t.Error("kv.db should not exist after migration")
		}
		if _, err := os.Stat(path.Join(dataDir, "kv.db.backup")); err != nil {
			t.Errorf("Backup file should exist after migration: %v", err)
		}

		_, err = runKVWithEnv(t, []string{"KV_DB=kv"}, "set", "extra-key", "extra-value")
		if err != nil {
			t.Fatalf("Failed to re-create kv.db: %v", err)
		}

		output, err := RunKV(t, "get", "nonexistent")
		_ = err
		if strings.Contains(output, "You have both") {
			t.Error("Warning should not re-appear after migration metadata was set")
		}
	})

	t.Run("warns once when both kv.db and default.db exist", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		dataDir := path.Join(tmpDir, "kv")

		_, err := runKVWithEnv(t, []string{"KV_DB=kv"}, "set", "some-key", "some-value")
		if err != nil {
			t.Fatalf("Failed to seed kv.db: %v", err)
		}

		kvPath := path.Join(dataDir, "kv.db")
		defaultPath := path.Join(dataDir, "default.db")
		dbBytes, err := os.ReadFile(kvPath)
		if err != nil {
			t.Fatalf("Failed to read kv.db: %v", err)
		}
		if err := os.WriteFile(defaultPath, dbBytes, 0o644); err != nil {
			t.Fatalf("Failed to create default.db: %v", err)
		}

		output, err := RunKV(t, "get", "nonexistent")
		_ = err
		if !strings.Contains(output, "You have both") {
			t.Error("Expected warning about both DBs existing")
		}

		output, err = RunKV(t, "get", "nonexistent")
		_ = err
		if strings.Contains(output, "You have both") {
			t.Error("Warning should only appear once")
		}
	})
}

func TestMigrateInvalidKeys(t *testing.T) {
	t.Run("no warning when no keys contain @", func(t *testing.T) {
		SetupTestDB(t)

		output, err := RunKV(t, "get", "nonexistent")
		_ = err
		if strings.Contains(output, "keys have '@'") {
			t.Error("There should be no warning with no @ keys")
		}
	})

	t.Run("warns once when keys contain @", func(t *testing.T) {
		SetupTestDB(t)

		_, err := runKVWithEnv(t, []string{"KV_NO_PARSE_KEYS=1"}, "set", "bad@key", "value")
		if err != nil {
			t.Fatalf("Failed to set key with @: %v", err)
		}

		output, err := RunKV(t, "get", "nonexistent")
		_ = err
		if !strings.Contains(output, "keys have '@'") {
			t.Errorf("Expected warning about @ keys, got: %s", output)
		}

		output, err = RunKV(t, "get", "nonexistent")
		_ = err
		if strings.Contains(output, "keys have '@'") {
			t.Error("Warning should only appear once")
		}
	})
}
