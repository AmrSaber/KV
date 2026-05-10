package tests

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDBSetNameCommand(t *testing.T) {
	t.Run("rename DB", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "x@mydb", "1")

		before := RunKVSuccess(t, "db", "list", "--output", "json")
		if !strings.Contains(before, "mydb") {
			t.Fatal("mydb should exist before rename")
		}

		RunKVSuccess(t, "db", "set", "name", "mydb", "renamed")

		after := RunKVSuccess(t, "db", "list", "--output", "json")
		if strings.Contains(after, "mydb") {
			t.Error("old name should not appear after rename")
		}
		if !strings.Contains(after, "renamed") {
			t.Error("new name should appear after rename")
		}
	})

	t.Run("data accessible after rename", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "x@mydb", "42")

		RunKVSuccess(t, "db", "set", "name", "mydb", "renamed")

		if got := RunKVSuccess(t, "get", "x@renamed"); got != "42" {
			t.Errorf("Expected '42', got: %s", got)
		}
	})

	t.Run("rename non-existent DB fails", func(t *testing.T) {
		SetupTestDB(t)

		output := RunKVFailure(t, "db", "set", "name", "nonexistent", "new")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist', got: %s", output)
		}
	})

	t.Run("rename to already existing name fails", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "x@db1", "1")
		RunKVSuccess(t, "set", "y@db2", "2")

		output := RunKVFailure(t, "db", "set", "name", "db1", "db2")
		if !strings.Contains(output, "already exists") {
			t.Errorf("Expected 'already exists', got: %s", output)
		}
	})
}

func TestDBSetDirectoryCommand(t *testing.T) {
	t.Run("change DB directory", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		newDir := tmpDir + "/new-db-location"
		RunKVSuccess(t, "set", "x@mydb", "1")

		RunKVSuccess(t, "db", "set", "directory", "mydb", newDir)

		output := RunKVSuccess(t, "db", "list", "--output", "json")
		var items []map[string]string
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		for _, item := range items {
			if item["name"] == "mydb" {
				if item["directory"] != newDir {
					t.Errorf("Expected directory %q, got: %q", newDir, item["directory"])
				}
				return
			}
		}
		t.Error("mydb should appear after directory change")
	})

	t.Run("data accessible after directory change", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		newDir := tmpDir + "/other-location"
		RunKVSuccess(t, "set", "x@mydb", "99")

		RunKVSuccess(t, "db", "set", "directory", "mydb", newDir)

		if got := RunKVSuccess(t, "get", "x@mydb"); got != "99" {
			t.Errorf("Expected '99', got: %s", got)
		}
	})
}
