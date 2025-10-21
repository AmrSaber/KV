package tests

import (
	"strings"
	"testing"
)

func TestRenameCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("basic rename", func(t *testing.T) {
		// Create a key
		RunKVSuccess(t, "set", "old-key", "test value")

		// Rename it
		RunKVSuccess(t, "rename", "old-key", "new-key")

		// Old key should not exist
		output := RunKVFailure(t, "get", "old-key")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}

		// New key should have the value
		output = RunKVSuccess(t, "get", "new-key")
		if output != "test value" {
			t.Errorf("Expected 'test value', got: %s", output)
		}
	})

	t.Run("rename preserves history", func(t *testing.T) {
		// Create a key with multiple values
		RunKVSuccess(t, "set", "history-key", "value1")
		RunKVSuccess(t, "set", "history-key", "value2")
		RunKVSuccess(t, "set", "history-key", "value3")

		// Rename it
		RunKVSuccess(t, "rename", "history-key", "renamed-history-key")

		// Check history is preserved
		output := RunKVSuccess(t, "history", "list", "renamed-history-key")
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

	t.Run("rename preserves encryption", func(t *testing.T) {
		// Create encrypted key
		RunKVSuccess(t, "set", "encrypted-key", "secret", "--password", "mypass")

		// Rename it
		RunKVSuccess(t, "rename", "encrypted-key", "renamed-encrypted")

		// Should be locked
		output := RunKVSuccess(t, "list", "renamed-encrypted")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Renamed key should be locked")
		}

		// Should decrypt with same password
		output = RunKVSuccess(t, "get", "renamed-encrypted", "--password", "mypass")
		if output != "secret" {
			t.Errorf("Expected 'secret', got: %s", output)
		}
	})

	t.Run("rename preserves TTL", func(t *testing.T) {
		// Create key with TTL
		RunKVSuccess(t, "set", "ttl-key", "temp", "--expires-after", "1h")

		// Rename it
		RunKVSuccess(t, "rename", "ttl-key", "renamed-ttl")

		// Should still have TTL
		output := RunKVSuccess(t, "ttl", "renamed-ttl")
		if !strings.Contains(output, "expires at") {
			t.Errorf("Renamed key should have TTL, got: %s", output)
		}
	})

	t.Run("rename non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "rename", "non-existent", "new-name")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("rename to existing key fails", func(t *testing.T) {
		// Create two keys
		RunKVSuccess(t, "set", "key1", "value1")
		RunKVSuccess(t, "set", "key2", "value2")

		// Try to rename key1 to key2
		output := RunKVFailure(t, "rename", "key1", "key2")
		if !strings.Contains(output, "already exists") {
			t.Errorf("Expected 'already exists' error, got: %s", output)
		}
	})
}
