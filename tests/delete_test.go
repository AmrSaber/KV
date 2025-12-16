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

		RunKVSuccess(t, "delete", "temp.", "--prefix")

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

func TestDeleteMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("delete multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "d1", "value1")
		RunKVSuccess(t, "set", "d2", "value2")
		RunKVSuccess(t, "set", "d3", "value3")

		RunKVSuccess(t, "delete", "d1", "d2", "d3")

		// All keys should be deleted
		output := RunKVFailure(t, "get", "d1")
		if !strings.Contains(output, "does not exist") {
			t.Error("d1 should be deleted")
		}

		output = RunKVFailure(t, "get", "d2")
		if !strings.Contains(output, "does not exist") {
			t.Error("d2 should be deleted")
		}

		output = RunKVFailure(t, "get", "d3")
		if !strings.Contains(output, "does not exist") {
			t.Error("d3 should be deleted")
		}
	})

	t.Run("delete multiple keys with prune", func(t *testing.T) {
		RunKVSuccess(t, "set", "p1", "value1")
		RunKVSuccess(t, "set", "p2", "value2")
		RunKVSuccess(t, "set", "p3", "value3")

		RunKVSuccess(t, "delete", "p1", "p2", "p3", "--prune")

		// All keys and their history should be deleted
		output := RunKVFailure(t, "history", "list", "p1")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Error("p1 history should be deleted")
		}

		output = RunKVFailure(t, "history", "list", "p2")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Error("p2 history should be deleted")
		}

		output = RunKVFailure(t, "history", "list", "p3")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Error("p3 history should be deleted")
		}
	})

	// Setup keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a")
	RunKVSuccess(t, "set", "b", "value-b")

	testCases := []struct {
		name string
		keys []string
	}{
		{"delete fails on first non-existent key", []string{"missing", "a", "b"}},
		{"delete fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"delete fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"delete"}, tc.keys...)
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should still exist (transaction rollback)
			output = RunKVSuccess(t, "get", "a")
			if output != "value-a" {
				t.Error("a should still exist due to transaction rollback")
			}

			output = RunKVSuccess(t, "get", "b")
			if output != "value-b" {
				t.Error("b should still exist due to transaction rollback")
			}
		})
	}
}
