package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestListCommand(t *testing.T) {
	t.Run("list all keys", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "key1", "value1")
		RunKVSuccess(t, "set", "key2", "value2")
		RunKVSuccess(t, "set", "key3", "value3")

		output := RunKVSuccess(t, "list")
		if !strings.Contains(output, "key1") {
			t.Error("Should contain key1")
		}
		if !strings.Contains(output, "key2") {
			t.Error("Should contain key2")
		}
		if !strings.Contains(output, "key3") {
			t.Error("Should contain key3")
		}
	})

	t.Run("list hides values without --values flag", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "key", "secret-value")

		output := RunKVSuccess(t, "list")
		if strings.Contains(output, "secret-value") {
			t.Error("Should not show values without --values flag")
		}
		if !strings.Contains(output, "key") {
			t.Error("Should show key names")
		}
	})

	t.Run("list --values shows values", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "key", "visible-value")

		output := RunKVSuccess(t, "list", "--values")
		if !strings.Contains(output, "visible-value") {
			t.Error("Should show values with --values flag")
		}
		if !strings.Contains(output, "key") {
			t.Error("Should show key names")
		}
	})

	t.Run("list with prefix", func(t *testing.T) {
		SetupTestDB(t)
		// Cover matching, same-namespace, and unrelated keys to validate filtering boundaries
		RunKVSuccess(t, "set", "app.db.host", "localhost")
		RunKVSuccess(t, "set", "app.db.port", "5432")
		RunKVSuccess(t, "set", "app.name", "myapp")
		RunKVSuccess(t, "set", "other", "value")

		output := RunKVSuccess(t, "list", "app.db")
		if !strings.Contains(output, "app.db.host") {
			t.Error("Should contain app.db.host")
		}
		if !strings.Contains(output, "app.db.port") {
			t.Error("Should contain app.db.port")
		}
		if strings.Contains(output, "app.name") {
			t.Error("Should not contain app.name")
		}
		if strings.Contains(output, "other") {
			t.Error("Should not contain other")
		}
	})

	t.Run("list shows locked keys", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "plain", "data")
		RunKVSuccess(t, "set", "encrypted", "secret", "--password=pass")

		output := RunKVSuccess(t, "list", "--values")
		if !strings.Contains(output, "plain") {
			t.Error("Should contain plain key")
		}
		if !strings.Contains(output, "encrypted") {
			t.Error("Should contain encrypted key")
		}
		if !strings.Contains(output, "[Locked]") {
			t.Error("Should show [Locked] for encrypted key")
		}
	})

	t.Run("list shows TTL column when keys have expiration", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "permanent", "data")
		RunKVSuccess(t, "set", "temporary", "data", "--expires-after", "1h")

		output := RunKVSuccess(t, "list")
		if !strings.Contains(output, "EXPIRES AT") {
			t.Error("Should show EXPIRES AT column")
		}
	})

	t.Run("list hides values by default", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "secret", "hidden-value")
		RunKVSuccess(t, "hide", "secret")

		output := RunKVSuccess(t, "list")
		if strings.Contains(output, "hidden-value") {
			t.Error("Should not show value of hidden key")
		}
	})

	t.Run("list --show reveals hidden values", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "secret", "hidden-value")
		RunKVSuccess(t, "hide", "secret")

		output := RunKVSuccess(t, "list", "--values", "--show")
		if !strings.Contains(output, "hidden-value") {
			t.Error("Should show value of hidden key with --show flag")
		}
		if strings.Contains(output, "[Hidden]") {
			t.Error("Should not show [Hidden] marker when --show is set")
		}
	})

	t.Run("list with JSON output", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "json-key", "json-value")

		output := RunKVSuccess(t, "list", "--output", "json", "--values")

		// Validate well-formed JSON with correct schema
		var items []map[string]any
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		found := false
		for _, item := range items {
			if item["key"] == "json-key" {
				found = true
				if item["value"] != "json-value" {
					t.Errorf("Expected value 'json-value', got: %v", item["value"])
				}
				break
			}
		}
		if !found {
			t.Error("JSON output should contain json-key")
		}
	})

	t.Run("list with YAML output", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "yaml-key", "yaml-value")

		output := RunKVSuccess(t, "list", "--output", "yaml", "--values")

		// Validate well-formed YAML with correct schema
		var items []map[string]any
		if err := yaml.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid YAML: %v", err)
		}

		found := false
		for _, item := range items {
			if item["key"] == "yaml-key" {
				found = true
				if item["value"] != "yaml-value" {
					t.Errorf("Expected value 'yaml-value', got: %v", item["value"])
				}
				break
			}
		}
		if !found {
			t.Error("YAML output should contain yaml-key")
		}
	})

	t.Run("list --deleted shows deleted keys", func(t *testing.T) {
		SetupTestDB(t)
		// Mix of active and deleted keys to verify --deleted filters correctly in both directions
		RunKVSuccess(t, "set", "active1", "value1")
		RunKVSuccess(t, "set", "active2", "value2")
		RunKVSuccess(t, "set", "deleted1", "value3")
		RunKVSuccess(t, "set", "deleted2", "value4")
		RunKVSuccess(t, "delete", "deleted1")
		RunKVSuccess(t, "delete", "deleted2")

		output := RunKVSuccess(t, "list", "--deleted")
		if !strings.Contains(output, "deleted1") {
			t.Error("Should show deleted1")
		}
		if !strings.Contains(output, "deleted2") {
			t.Error("Should show deleted2")
		}
		if strings.Contains(output, "active1") {
			t.Error("Should not show active1")
		}
		if strings.Contains(output, "active2") {
			t.Error("Should not show active2")
		}
	})

	t.Run("list --deleted hides values", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "gone", "secret-value")
		RunKVSuccess(t, "delete", "gone")

		output := RunKVSuccess(t, "list", "--deleted")
		if strings.Contains(output, "secret-value") {
			t.Error("Should not show values for deleted keys")
		}
	})

	t.Run("list --all shows items from all DBs sorted by DB then key", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "a@mydb", "val_a")
		RunKVSuccess(t, "set", "z@mydb", "val_z")
		RunKVSuccess(t, "set", "m_default", "val_m")
		RunKVSuccess(t, "set", "a_default", "val_a")

		output := RunKVSuccess(t, "list", "--all", "--output", "json")

		var items []map[string]any
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		if len(items) != 4 {
			t.Fatalf("Expected 4 items, got %d", len(items))
		}

		// Verify sorting: DB first, then key
		for _, item := range items {
			if _, ok := item["db"]; !ok {
				t.Errorf("Item %v should have 'db' field when --all is set", item["key"])
			}
		}

		expected := []struct{ db, key string }{
			{"default", "a_default"},
			{"default", "m_default"},
			{"mydb", "a"},
			{"mydb", "z"},
		}
		for i, exp := range expected {
			if items[i]["db"] != exp.db {
				t.Errorf("Item %d: expected db %q, got %v", i, exp.db, items[i]["db"])
			}
			if items[i]["key"] != exp.key {
				t.Errorf("Item %d: expected key %q, got %v", i, exp.key, items[i]["key"])
			}
		}
	})

	t.Run("list --all displays DB column in table view", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "k1@mydb", "v1")
		RunKVSuccess(t, "set", "k2", "v2")

		output := RunKVSuccess(t, "list", "--all")

		if !strings.Contains(output, "DB") {
			t.Error("Should show DB column header")
		}
		if !strings.Contains(output, "mydb") {
			t.Error("Should show mydb DB name")
		}
		if !strings.Contains(output, "default") {
			t.Error("Should show default DB name")
		}
	})

	t.Run("list --all with prefix filters across DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "app.host@mydb", "host")
		RunKVSuccess(t, "set", "app.port@mydb", "port")
		RunKVSuccess(t, "set", "other@mydb", "other")
		RunKVSuccess(t, "set", "app.name", "name")

		output := RunKVSuccess(t, "list", "--all", "app", "--output", "json")

		var items []map[string]any
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		if len(items) != 3 {
			t.Fatalf("Expected 3 items, got %d", len(items))
		}

		for _, item := range items {
			key, ok := item["key"].(string)
			if !ok || !strings.HasPrefix(key, "app") {
				t.Errorf("Expected key with 'app' prefix, got %v", item["key"])
			}
		}
	})

	t.Run("list --all --deleted shows deleted items from all DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "gone@mydb", "secret")
		RunKVSuccess(t, "set", "stay@mydb", "visible")
		RunKVSuccess(t, "set", "deleted_default", "deleted_val")
		RunKVSuccess(t, "set", "active", "active_val")
		RunKVSuccess(t, "delete", "gone@mydb")
		RunKVSuccess(t, "delete", "deleted_default")

		output := RunKVSuccess(t, "list", "--all", "--deleted", "--output", "json")

		var items []map[string]any
		if err := json.Unmarshal([]byte(output), &items); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		if len(items) != 2 {
			t.Fatalf("Expected 2 deleted items, got %d", len(items))
		}

		for _, item := range items {
			if _, hasValue := item["value"]; hasValue {
				t.Errorf("Deleted item %v should not have a value", item["key"])
			}
		}
	})
}

func TestListCompletions(t *testing.T) {
	t.Run("completions match substring in the middle of a key", func(t *testing.T) {
		SetupTestDB(t)
		// Cover substring match from the middle vs prefix-only and unrelated keys
		RunKVSuccess(t, "set", "app.db.host", "localhost")
		RunKVSuccess(t, "set", "app.db.port", "5432")
		RunKVSuccess(t, "set", "app.name", "myapp")

		output := RunKVSuccess(t, "__complete", "list", "db")
		if !strings.Contains(output, "app.db.host") {
			t.Error("Should complete app.db.host")
		}
		if !strings.Contains(output, "app.db.port") {
			t.Error("Should complete app.db.port")
		}
		if strings.Contains(output, "app.name") {
			t.Error("Should not complete app.name")
		}
	})
}
