package tests

import (
	"strings"
	"testing"
)

func TestCopyCrossDB(t *testing.T) {
	t.Run("basic copy across DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "src@db1", "value")

		// Same-DB copy
		RunKVSuccess(t, "copy", "src@db1", "same-db@db1")
		if got := RunKVSuccess(t, "get", "same-db@db1"); got != "value" {
			t.Errorf("Same-DB copy: expected 'value', got: %s", got)
		}

		// Cross-DB copy
		RunKVSuccess(t, "copy", "src@db1", "dst@db2")
		if got := RunKVSuccess(t, "get", "dst@db2"); got != "value" {
			t.Errorf("Cross-DB copy: expected 'value', got: %s", got)
		}

		// Source stays intact
		if got := RunKVSuccess(t, "get", "src@db1"); got != "value" {
			t.Errorf("Source should be unchanged, got: %s", got)
		}

		// @ shorthand
		RunKVSuccess(t, "copy", "src@db1", "@db3")
		if got := RunKVSuccess(t, "get", "src@db3"); got != "value" {
			t.Errorf("@ shorthand: expected 'value', got: %s", got)
		}
	})

	t.Run("copy preserves lock+hide but not TTL", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "src@db1", "data", "--password=pass", "--expires-after", "1h")
		RunKVSuccess(t, "hide", "src@db1")

		RunKVSuccess(t, "copy", "src@db1", "dst@db2")

		// Lock preserves
		listOut := RunKVSuccess(t, "list", "dst@db2")
		if !strings.Contains(listOut, "[Locked]") {
			t.Error("Lock state should be preserved")
		}
		if got := RunKVSuccess(t, "get", "dst@db2", "--password=pass"); got != "data" {
			t.Errorf("Decrypted value: expected 'data', got: %s", got)
		}

		// Hide preserves (check via JSON since [Locked] takes precedence in table view)
		if !strings.Contains(RunKVSuccess(t, "list", "dst@db2", "--output", "json"), `"isHidden": true`) {
			t.Error("Hidden state should be preserved")
		}

		// TTL does NOT preserve
		ttlOut, _ := RunKV(t, "ttl", "dst@db2")
		if !strings.Contains(ttlOut, "does not expire") {
			t.Errorf("TTL should not be preserved, got: %s", ttlOut)
		}
	})

	t.Run("copy non-existent key fails", func(t *testing.T) {
		SetupTestDB(t)
		output := RunKVFailure(t, "copy", "nonexistent@db1", "dst@db2")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist', got: %s", output)
		}
	})
}

func TestMoveCrossDB(t *testing.T) {
	t.Run("basic move across DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "k@db1", "value")

		// Same-DB move
		RunKVSuccess(t, "move", "k@db1", "same-db@db1")
		RunKVFailure(t, "get", "k@db1")
		if got := RunKVSuccess(t, "get", "same-db@db1"); got != "value" {
			t.Errorf("Same-DB: expected 'value', got: %s", got)
		}

		// Cross-DB move
		RunKVSuccess(t, "move", "same-db@db1", "moved@db2")
		RunKVFailure(t, "get", "same-db@db1")
		if got := RunKVSuccess(t, "get", "moved@db2"); got != "value" {
			t.Errorf("Cross-DB: expected 'value', got: %s", got)
		}

		// @ shorthand
		RunKVSuccess(t, "copy", "moved@db2", "tmp@db2")
		RunKVSuccess(t, "move", "tmp@db2", "@db3")
		RunKVFailure(t, "get", "tmp@db2")
		if got := RunKVSuccess(t, "get", "tmp@db3"); got != "value" {
			t.Errorf("@ shorthand: expected 'value', got: %s", got)
		}
	})

	t.Run("move preserves history+lock+TTL across DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "k@db1", "v1")
		RunKVSuccess(t, "set", "k@db1", "v2")
		RunKVSuccess(t, "set", "k@db1", "v3", "--password=pass", "--expires-after", "30m")

		RunKVSuccess(t, "move", "k@db1", "k@db2")

		histOut := RunKVSuccess(t, "history", "list", "k@db2")
		if !strings.Contains(histOut, "v1") || !strings.Contains(histOut, "v2") {
			t.Error("History should be preserved")
		}

		if got := RunKVSuccess(t, "get", "k@db2", "--password=pass"); got != "v3" {
			t.Errorf("Encrypted value: expected 'v3', got: %s", got)
		}

		ttlOut := RunKVSuccess(t, "ttl", "k@db2")
		if !strings.Contains(ttlOut, "expires at") {
			t.Errorf("TTL should be preserved, got: %s", ttlOut)
		}
	})

	t.Run("move to existing key fails across DBs", func(t *testing.T) {
		SetupTestDB(t)
		RunKVSuccess(t, "set", "existing@db2", "occupied")
		RunKVSuccess(t, "set", "source@db1", "value")
		output := RunKVFailure(t, "move", "source@db1", "existing@db2")
		if !strings.Contains(output, "already exists") {
			t.Errorf("Expected 'already exists', got: %s", output)
		}
	})
}
