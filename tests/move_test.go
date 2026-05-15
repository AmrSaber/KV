package tests

import (
	"strings"
	"testing"
)

func TestMoveCommand(t *testing.T) {
	t.Run("basic move", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "old-key", "test value")
		RunKVSuccess(t, "move", "old-key", "new-key")

		output := RunKVFailure(t, "get", "old-key")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}

		output = RunKVSuccess(t, "get", "new-key")
		if output != "test value" {
			t.Errorf("Expected 'test value', got: %s", output)
		}
	})

	t.Run("move preserves history", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "history-key", "value1")
		RunKVSuccess(t, "set", "history-key", "value2")
		RunKVSuccess(t, "set", "history-key", "value3")
		RunKVSuccess(t, "move", "history-key", "moved-history-key")

		output := RunKVSuccess(t, "history", "list", "moved-history-key")
		if !strings.Contains(output, "value1") {
			t.Error("History should contain value1")
		}
		if !strings.Contains(output, "value2") {
			t.Error("History should contain value2")
		}
		if !strings.Contains(output, "value3") {
			t.Error("History should contain value3")
		}
	})

	t.Run("move preserves encryption", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "encrypted-key", "secret", "--password=mypass")
		RunKVSuccess(t, "move", "encrypted-key", "moved-encrypted")

		output := RunKVSuccess(t, "list", "moved-encrypted", "--values")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Moved key should be locked")
		}

		output = RunKVSuccess(t, "get", "moved-encrypted", "--password=mypass")
		if output != "secret" {
			t.Errorf("Expected 'secret', got: %s", output)
		}
	})

	t.Run("move preserves TTL", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "ttl-key", "temp", "--expires-after", "1h")
		RunKVSuccess(t, "move", "ttl-key", "moved-ttl")

		output := RunKVSuccess(t, "ttl", "moved-ttl")
		if !strings.Contains(output, "expires at") {
			t.Errorf("Moved key should have TTL, got: %s", output)
		}
	})

	t.Run("move non-existent key fails", func(t *testing.T) {
		SetupTestDB(t)
		output := RunKVFailure(t, "move", "non-existent", "new-name")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("move to existing key fails", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "key1", "value1")
		RunKVSuccess(t, "set", "key2", "value2")
		output := RunKVFailure(t, "move", "key1", "key2")
		if !strings.Contains(output, "already exists") {
			t.Errorf("Expected 'already exists' error, got: %s", output)
		}
	})
}
