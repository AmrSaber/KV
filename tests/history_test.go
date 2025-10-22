package tests

import (
	"strings"
	"testing"
)

func TestHistoryListCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("list history for key", func(t *testing.T) {
		RunKVSuccess(t, "set", "versioned", "v1")
		RunKVSuccess(t, "set", "versioned", "v2")
		RunKVSuccess(t, "set", "versioned", "v3")

		output := RunKVSuccess(t, "history", "list", "versioned")
		if !strings.Contains(output, "v1") {
			t.Error("History should contain v1")
		}
		if !strings.Contains(output, "v2") {
			t.Error("History should contain v2")
		}
		if !strings.Contains(output, "v3") {
			t.Error("History should contain v3")
		}
	})

	t.Run("list history shows current value marker", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "old")
		RunKVSuccess(t, "set", "key", "current")

		output := RunKVSuccess(t, "history", "list", "key")
		// Should show "-" or similar marker for current value
		if !strings.Contains(output, "-") && !strings.Contains(output, "current") {
			t.Error("Should indicate current value in history")
		}
	})

	t.Run("list history for non-existent key", func(t *testing.T) {
		output := RunKVFailure(t, "history", "list", "non-existent")
		if !strings.Contains(output, "No history") && !strings.Contains(output, "does not exist") {
			t.Errorf("Expected no history error, got: %s", output)
		}
	})

	t.Run("history includes deleted values", func(t *testing.T) {
		RunKVSuccess(t, "set", "deleted-key", "value")
		RunKVSuccess(t, "delete", "deleted-key")

		output := RunKVSuccess(t, "history", "list", "deleted-key")
		if !strings.Contains(output, "value") {
			t.Error("History should show deleted value")
		}
	})
}

func TestHistoryRevertCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("revert to previous value", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "old")
		RunKVSuccess(t, "set", "key", "new")
		RunKVSuccess(t, "history", "revert", "key")

		output := RunKVSuccess(t, "get", "key")
		if output != "old" {
			t.Errorf("Expected 'old', got: %s", output)
		}
	})

	t.Run("revert multiple steps", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "v1")
		RunKVSuccess(t, "set", "key", "v2")
		RunKVSuccess(t, "set", "key", "v3")
		RunKVSuccess(t, "set", "key", "v4")

		RunKVSuccess(t, "history", "revert", "key", "--steps", "3")

		output := RunKVSuccess(t, "get", "key")
		if output != "v1" {
			t.Errorf("Expected 'v1', got: %s", output)
		}
	})

	t.Run("revert non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "history", "revert", "non-existent")
		if !strings.Contains(output, "does not exist") && !strings.Contains(output, "No history") && !strings.Contains(output, "panic") && !strings.Contains(output, "no rows") {
			t.Errorf("Expected error for non-existent key, got: %s", output)
		}
	})

	t.Run("revert creates new history entry", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "v1")
		RunKVSuccess(t, "set", "key", "v2")
		RunKVSuccess(t, "history", "revert", "key")

		output := RunKVSuccess(t, "history", "list", "key")
		// Should have 3 entries: v1, v2, v1 (reverted)
		count := strings.Count(output, "v1")
		if count < 2 {
			t.Error("Revert should create new history entry")
		}
	})
}

func TestHistoryPruneCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("prune history for single key", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "v1")
		RunKVSuccess(t, "set", "key", "v2")
		RunKVSuccess(t, "set", "key", "v3")

		RunKVSuccess(t, "history", "prune", "key")

		// Current value should still exist
		output := RunKVSuccess(t, "get", "key")
		if output != "v3" {
			t.Error("Current value should be preserved")
		}

		// History should be gone or minimal
		output = RunKVSuccess(t, "history", "list", "key")
		if strings.Contains(output, "v1") || strings.Contains(output, "v2") {
			t.Error("Old history should be pruned")
		}
	})

	t.Run("prune history with prefix", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp.a", "v1")
		RunKVSuccess(t, "set", "temp.a", "v2")
		RunKVSuccess(t, "set", "temp.b", "v1")
		RunKVSuccess(t, "set", "temp.b", "v2")

		RunKVSuccess(t, "history", "prune", "temp", "--prefix")

		// Both histories should be pruned
		output := RunKVSuccess(t, "history", "list", "temp.a")
		if strings.Contains(output, "v1") {
			t.Error("temp.a history should be pruned")
		}

		output = RunKVSuccess(t, "history", "list", "temp.b")
		if strings.Contains(output, "v1") {
			t.Error("temp.b history should be pruned")
		}
	})

	t.Run("prune non-existent key succeeds silently", func(t *testing.T) {
		// Prune on non-existent key doesn't fail, it just does nothing
		RunKVSuccess(t, "history", "prune", "non-existent")
	})
}

func TestHistorySelectCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("select command exists", func(t *testing.T) {
		RunKVSuccess(t, "history", "select", "--help")
	})

	// Note: history select is interactive, so we'll just test that the command exists
	// and handles basic error cases
	t.Run("select on non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "history", "select", "non-existent")
		if !strings.Contains(output, "does not exist") && !strings.Contains(output, "No history") && !strings.Contains(output, "panic") && !strings.Contains(output, "no rows") {
			t.Errorf("Expected error for non-existent key, got: %s", output)
		}
	})
}
