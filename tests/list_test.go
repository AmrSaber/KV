package tests

import (
	"strings"
	"testing"
)

func TestListCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("list all keys", func(t *testing.T) {
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
		RunKVSuccess(t, "set", "permanent", "data")
		RunKVSuccess(t, "set", "temporary", "data", "--expires-after", "1h")

		output := RunKVSuccess(t, "list")
		if !strings.Contains(output, "EXPIRES AT") {
			t.Error("Should show EXPIRES AT column")
		}
	})

	t.Run("list with no-values flag", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "secret-value")

		output := RunKVSuccess(t, "list", "--no-values")
		if strings.Contains(output, "secret-value") {
			t.Error("Should not show values with --no-values flag")
		}
		if !strings.Contains(output, "key") {
			t.Error("Should show key names")
		}
	})

	t.Run("list with JSON output", func(t *testing.T) {
		RunKVSuccess(t, "set", "json-key", "json-value")

		output := RunKVSuccess(t, "list", "--output", "json")
		if !strings.Contains(output, "json-key") {
			t.Error("JSON should contain key")
		}
		if !strings.Contains(output, "json-value") {
			t.Error("JSON should contain value")
		}
		if !strings.HasPrefix(output, "[") {
			t.Error("JSON output should start with [")
		}
	})
}
