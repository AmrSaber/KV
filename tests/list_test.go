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
		RunKVSuccess(t, "set", "encrypted", "secret", "--password", "pass")

		output := RunKVSuccess(t, "list")
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

	t.Run("list with no-values flag", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "key", "secret-value")

		output := RunKVSuccess(t, "list", "--no-values")
		if strings.Contains(output, "secret-value") {
			t.Error("Should not show values with --no-values flag")
		}
		if !strings.Contains(output, "key") {
			t.Error("Should show key names")
		}
	})

	t.Run("list hides hidden values by default", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "secret", "hidden-value")
		RunKVSuccess(t, "hide", "secret")

		output := RunKVSuccess(t, "list")
		if strings.Contains(output, "hidden-value") {
			t.Error("Should not show value of hidden key")
		}
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Should show [Hidden] marker for hidden key")
		}
	})

	t.Run("list --show reveals hidden values", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "secret", "hidden-value")
		RunKVSuccess(t, "hide", "secret")

		output := RunKVSuccess(t, "list", "--show")
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

		output := RunKVSuccess(t, "list", "--output", "json")

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

		output := RunKVSuccess(t, "list", "--output", "yaml")

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
