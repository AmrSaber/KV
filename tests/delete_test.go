package tests

import (
	"strings"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("delete existing key", func(t *testing.T) {
		RunKVSuccess(t, "set", "to-delete", "value")
		RunKVSuccess(t, "delete", "to-delete")

		output := RunKVFailure(t, "get", "to-delete")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Key should be deleted, got: %s", output)
		}
	})

	t.Run("delete keeps history by default", func(t *testing.T) {
		RunKVSuccess(t, "set", "hist-key", "value1")
		RunKVSuccess(t, "set", "hist-key", "value2")
		RunKVSuccess(t, "delete", "hist-key")

		// History should still exist
		output := RunKVSuccess(t, "history", "list", "hist-key")
		if !strings.Contains(output, "value1") || !strings.Contains(output, "value2") {
			t.Error("History should be preserved after soft delete")
		}
	})

	t.Run("delete with prune removes history", func(t *testing.T) {
		RunKVSuccess(t, "set", "prune-key", "value")
		RunKVSuccess(t, "delete", "prune-key", "--prune")

		// History should not exist
		output := RunKVFailure(t, "history", "list", "prune-key")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Errorf("History should be removed with --prune, got: %s", output)
		}
	})

	t.Run("delete non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "delete", "non-existent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("delete with prefix", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp.key1", "value1")
		RunKVSuccess(t, "set", "temp.key2", "value2")
		RunKVSuccess(t, "set", "keep.key", "value3")

		RunKVSuccess(t, "delete", "temp", "--prefix")

		// temp.* keys should be gone
		output := RunKVFailure(t, "get", "temp.key1")
		if !strings.Contains(output, "does not exist") {
			t.Error("temp.key1 should be deleted")
		}

		output = RunKVFailure(t, "get", "temp.key2")
		if !strings.Contains(output, "does not exist") {
			t.Error("temp.key2 should be deleted")
		}

		// keep.key should still exist
		output = RunKVSuccess(t, "get", "keep.key")
		if output != "value3" {
			t.Error("keep.key should still exist")
		}
	})

	t.Run("delete with prefix and prune", func(t *testing.T) {
		RunKVSuccess(t, "set", "cache.a", "value")
		RunKVSuccess(t, "set", "cache.b", "value")

		RunKVSuccess(t, "delete", "cache", "--prefix", "--prune")

		// History should be completely removed
		output := RunKVFailure(t, "history", "list", "cache.a")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Error("History should be removed for cache.a")
		}
	})

	t.Run("delete using del alias", func(t *testing.T) {
		RunKVSuccess(t, "set", "alias-test", "value")
		RunKVSuccess(t, "del", "alias-test")

		output := RunKVFailure(t, "get", "alias-test")
		if !strings.Contains(output, "does not exist") {
			t.Error("del alias should work")
		}
	})

	t.Run("delete using rm alias", func(t *testing.T) {
		RunKVSuccess(t, "set", "rm-test", "value")
		RunKVSuccess(t, "rm", "rm-test")

		output := RunKVFailure(t, "get", "rm-test")
		if !strings.Contains(output, "does not exist") {
			t.Error("rm alias should work")
		}
	})
}
