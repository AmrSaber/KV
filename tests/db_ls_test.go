package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestDBListCommand(t *testing.T) {
	t.Run("list default DB", func(t *testing.T) {
		SetupTestDB(t)

		output := RunKVSuccess(t, "db", "list")
		if !strings.Contains(output, "default") {
			t.Error("Should show default DB")
		}
	})

	t.Run("list all registered DBs", func(t *testing.T) {
		SetupTestDB(t)

		RunKVSuccess(t, "set", "x", "1", "--db", "adb")
		RunKVSuccess(t, "set", "y", "2", "--db", "zdb")
		RunKVSuccess(t, "set", "z", "3", "--db", "mdb")

		output := RunKVSuccess(t, "db", "list")
		for _, name := range []string{"adb", "default", "mdb", "zdb"} {
			if !strings.Contains(output, name) {
				t.Errorf("Should show %q DB", name)
			}
		}
	})

	t.Run("list with JSON output", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "x", "1", "--db", "json-db")

		output := RunKVSuccess(t, "db", "list", "--output", "json")

		var items []map[string]string
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		found := false
		for _, item := range items {
			if item["name"] == "json-db" {
				found = true
				if item["directory"] == "" {
					t.Error("Should have a directory path")
				}
				break
			}
		}
		if !found {
			t.Error("JSON output should contain json-db")
		}
	})

	t.Run("list with YAML output", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "x", "1", "--db", "yaml-db")

		output := RunKVSuccess(t, "db", "list", "--output", "yaml")

		var items []map[string]string
		if err := yaml.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid YAML: %v", err)
		}

		found := false
		for _, item := range items {
			if item["name"] == "yaml-db" {
				found = true
				if item["directory"] == "" {
					t.Error("Should have a directory path")
				}
				break
			}
		}
		if !found {
			t.Error("YAML output should contain yaml-db")
		}
	})

	t.Run("db list alias works", func(t *testing.T) {
		SetupTestDB(t)

		output := RunKVSuccess(t, "db", "ls")
		if !strings.Contains(output, "default") {
			t.Error("db ls alias should show default DB")
		}
	})

	t.Run("shows directory paths for all DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "db", "list")

		output := RunKVSuccess(t, "db", "list")
		if !strings.Contains(output, "default") {
			t.Error("Should show default DB name")
		}
		if !strings.Contains(output, "/") {
			t.Error("Should show directory path for each DB")
		}
	})
}
