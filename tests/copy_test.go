package tests

import (
	"strings"
	"testing"
)

func TestCopyCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("basic copy", func(t *testing.T) {
		// Create a key
		RunKVSuccess(t, "set", "source", "original value")

		// Copy it
		RunKVSuccess(t, "copy", "source", "destination")

		// Both should exist with same value
		sourceValue := RunKVSuccess(t, "get", "source")
		destValue := RunKVSuccess(t, "get", "destination")

		if sourceValue != "original value" {
			t.Errorf("Source should be 'original value', got: %s", sourceValue)
		}
		if destValue != "original value" {
			t.Errorf("Destination should be 'original value', got: %s", destValue)
		}
	})

	t.Run("copy to existing key overwrites", func(t *testing.T) {
		// Create two keys
		RunKVSuccess(t, "set", "src", "source value")
		RunKVSuccess(t, "set", "dst", "dest value")

		// Copy source to destination
		RunKVSuccess(t, "copy", "src", "dst")

		// Destination should have source value
		output := RunKVSuccess(t, "get", "dst")
		if output != "source value" {
			t.Errorf("Expected 'source value', got: %s", output)
		}
	})

	t.Run("copy preserves encryption", func(t *testing.T) {
		// Create encrypted key
		RunKVSuccess(t, "set", "encrypted-src", "secret data", "--password", "testpass")

		// Copy it
		RunKVSuccess(t, "copy", "encrypted-src", "encrypted-dst")

		// Destination should also be locked
		output := RunKVSuccess(t, "list", "encrypted-dst")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Copied key should be locked")
		}

		// Should decrypt with same password
		output = RunKVSuccess(t, "get", "encrypted-dst", "--password", "testpass")
		if output != "secret data" {
			t.Errorf("Expected 'secret data', got: %s", output)
		}
	})

	t.Run("copy does not preserve TTL", func(t *testing.T) {
		// Create key with TTL
		RunKVSuccess(t, "set", "ttl-src", "temp data", "--expires-after", "1h")

		// Verify source has TTL
		output := RunKVSuccess(t, "ttl", "ttl-src")
		if !strings.Contains(output, "expires at") {
			t.Error("Source should have TTL")
		}

		// Copy it
		RunKVSuccess(t, "copy", "ttl-src", "ttl-dst")

		// Destination should NOT have TTL (ttl command may return error for keys without TTL)
		output, _ = RunKV(t, "ttl", "ttl-dst")
		if !strings.Contains(output, "does not expire") {
			t.Errorf("Destination should not have TTL, got: %s", output)
		}
	})

	t.Run("copy non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "copy", "non-existent", "dest")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("copy creates history entry for existing destination", func(t *testing.T) {
		// Create destination with initial value
		RunKVSuccess(t, "set", "hist-dst", "old value")

		// Create source
		RunKVSuccess(t, "set", "hist-src", "new value")

		// Copy source to destination
		RunKVSuccess(t, "copy", "hist-src", "hist-dst")

		// Check history shows both values
		output := RunKVSuccess(t, "history", "list", "hist-dst")
		if !strings.Contains(output, "old value") {
			t.Error("History should contain old value")
		}
		if !strings.Contains(output, "new value") {
			t.Error("History should contain new value")
		}
	})
}
