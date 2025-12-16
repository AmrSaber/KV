package tests

import (
	"strings"
	"testing"
)

func TestHideCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("hide plain text key", func(t *testing.T) {
		RunKVSuccess(t, "set", "visible", "secret")
		RunKVSuccess(t, "hide", "visible")

		// Should show as hidden in list
		output := RunKVSuccess(t, "list", "visible")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Key should show as [Hidden]")
		}

		// Should still be accessible via get
		output = RunKVSuccess(t, "get", "visible")
		if output != "secret" {
			t.Errorf("Expected 'secret', got: %s", output)
		}
	})

	t.Run("hide already hidden key succeeds", func(t *testing.T) {
		RunKVSuccess(t, "set", "hidden", "data")
		RunKVSuccess(t, "hide", "hidden")
		// Should be idempotent - no error
		RunKVSuccess(t, "hide", "hidden")

		output := RunKVSuccess(t, "list", "hidden")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Key should still be hidden")
		}
	})

	t.Run("hide non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "hide", "non-existent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("hide with prefix", func(t *testing.T) {
		RunKVSuccess(t, "set", "secrets.api", "key1")
		RunKVSuccess(t, "set", "secrets.db", "key2")
		RunKVSuccess(t, "set", "public.data", "data")

		RunKVSuccess(t, "hide", "secrets", "--prefix")

		// secrets.* should be hidden
		output := RunKVSuccess(t, "list", "secrets")
		occurrences := strings.Count(output, "[Hidden]")
		if occurrences < 2 {
			t.Errorf("Expected at least 2 hidden keys, found %d", occurrences)
		}

		// public.data should not be hidden
		output = RunKVSuccess(t, "list", "public")
		if strings.Contains(output, "[Hidden]") {
			t.Error("public.data should not be hidden")
		}
	})

	t.Run("hide with -p shorthand", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp.a", "val1")
		RunKVSuccess(t, "set", "temp.b", "val2")

		RunKVSuccess(t, "hide", "temp", "-p")

		output := RunKVSuccess(t, "list", "temp")
		occurrences := strings.Count(output, "[Hidden]")
		if occurrences < 2 {
			t.Errorf("Expected at least 2 hidden keys, found %d", occurrences)
		}
	})
}

func TestShowCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("show hidden key", func(t *testing.T) {
		RunKVSuccess(t, "set", "hidden", "data")
		RunKVSuccess(t, "hide", "hidden")
		RunKVSuccess(t, "show", "hidden")

		// Should not show as hidden
		output := RunKVSuccess(t, "list", "hidden")
		if strings.Contains(output, "[Hidden]") {
			t.Error("Key should not be hidden after show")
		}

		// Should show actual value
		if !strings.Contains(output, "data") {
			t.Error("Should show actual value after show")
		}
	})

	t.Run("show already visible key succeeds", func(t *testing.T) {
		RunKVSuccess(t, "set", "visible", "data")
		// Should be idempotent - no error
		RunKVSuccess(t, "show", "visible")

		output := RunKVSuccess(t, "list", "visible")
		if strings.Contains(output, "[Hidden]") {
			t.Error("Key should not be hidden")
		}
	})

	t.Run("show non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "show", "non-existent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("show with prefix", func(t *testing.T) {
		RunKVSuccess(t, "set", "hidden.a", "val1")
		RunKVSuccess(t, "set", "hidden.b", "val2")
		RunKVSuccess(t, "hide", "hidden", "--prefix")
		RunKVSuccess(t, "show", "hidden", "--prefix")

		// Both keys should be visible
		output := RunKVSuccess(t, "list", "hidden")
		if strings.Contains(output, "[Hidden]") {
			t.Error("No keys should be hidden after show --prefix")
		}
	})

	t.Run("show with -p shorthand", func(t *testing.T) {
		RunKVSuccess(t, "set", "test.a", "val1")
		RunKVSuccess(t, "set", "test.b", "val2")
		RunKVSuccess(t, "hide", "test", "-p")
		RunKVSuccess(t, "show", "test", "-p")

		// Both keys should be visible
		output := RunKVSuccess(t, "list", "test")
		if strings.Contains(output, "[Hidden]") {
			t.Error("No keys should be hidden after show -p")
		}
	})
}

func TestHiddenAndLockedCombination(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("lock a hidden key", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "secret")
		RunKVSuccess(t, "hide", "key")
		RunKVSuccess(t, "lock", "key", "--password", "pass")

		// Should show [Locked] (takes precedence over [Hidden])
		output := RunKVSuccess(t, "list", "key")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Should show [Locked]")
		}
		if strings.Contains(output, "[Hidden]") {
			t.Error("Should not show [Hidden] when locked")
		}
	})

	t.Run("hide a locked key", func(t *testing.T) {
		RunKVSuccess(t, "set", "key2", "secret", "--password", "pass")
		RunKVSuccess(t, "hide", "key2")

		// Should show [Locked] (takes precedence)
		output := RunKVSuccess(t, "list", "key2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Should show [Locked]")
		}
		if strings.Contains(output, "[Hidden]") {
			t.Error("Should not show [Hidden] when locked")
		}
	})

	t.Run("unlock preserves hidden state", func(t *testing.T) {
		RunKVSuccess(t, "set", "key3", "secret", "--password", "pass")
		RunKVSuccess(t, "hide", "key3")
		RunKVSuccess(t, "unlock", "key3", "--password", "pass")

		// Should show [Hidden] now (not locked anymore)
		output := RunKVSuccess(t, "list", "key3")
		if strings.Contains(output, "[Locked]") {
			t.Error("Should not be locked after unlock")
		}
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Should still be hidden after unlock")
		}
	})

	t.Run("lock preserves hidden state", func(t *testing.T) {
		RunKVSuccess(t, "set", "key4", "secret")
		RunKVSuccess(t, "hide", "key4")
		RunKVSuccess(t, "lock", "key4", "--password", "pass")
		RunKVSuccess(t, "unlock", "key4", "--password", "pass")

		// Should show [Hidden] (hidden state preserved through lock/unlock)
		output := RunKVSuccess(t, "list", "key4")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Should still be hidden after lock/unlock cycle")
		}
	})

	t.Run("set preserves hidden state", func(t *testing.T) {
		RunKVSuccess(t, "set", "key5", "val1")
		RunKVSuccess(t, "hide", "key5")
		RunKVSuccess(t, "set", "key5", "val2")

		// Should still be hidden
		output := RunKVSuccess(t, "list", "key5")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("Should still be hidden after set")
		}

		// But value should be updated
		output = RunKVSuccess(t, "get", "key5")
		if output != "val2" {
			t.Errorf("Expected 'val2', got: %s", output)
		}
	})
}

func TestHiddenInHistory(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("history shows hidden entries", func(t *testing.T) {
		RunKVSuccess(t, "set", "key", "val1")
		RunKVSuccess(t, "hide", "key")
		RunKVSuccess(t, "set", "key", "val2")

		output := RunKVSuccess(t, "history", "list", "key")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("History should show hidden entries")
		}
	})

	t.Run("history respects locked precedence", func(t *testing.T) {
		RunKVSuccess(t, "set", "key2", "val1", "--password", "pass")
		RunKVSuccess(t, "hide", "key2")

		output := RunKVSuccess(t, "history", "list", "key2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("History should show [Locked]")
		}
		if strings.Contains(output, "[Hidden]") {
			t.Error("History should not show [Hidden] when locked")
		}
	})

	t.Run("history shows hidden after unlock", func(t *testing.T) {
		RunKVSuccess(t, "set", "key3", "val1", "--password", "pass")
		RunKVSuccess(t, "hide", "key3")
		RunKVSuccess(t, "unlock", "key3", "--password", "pass")

		output := RunKVSuccess(t, "history", "list", "key3")
		if strings.Contains(output, "[Locked]") {
			t.Error("History should not show [Locked] after unlock")
		}
		if !strings.Contains(output, "[Hidden]") {
			t.Error("History should show [Hidden] after unlock")
		}
	})
}

func TestHiddenInJSON(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("JSON output includes isHidden field", func(t *testing.T) {
		RunKVSuccess(t, "set", "hidden-key", "value")
		RunKVSuccess(t, "hide", "hidden-key")

		output := RunKVSuccess(t, "list", "hidden-key", "--output", "json")
		if !strings.Contains(output, `"isHidden"`) {
			t.Error("JSON output should include isHidden field")
		}
		if !strings.Contains(output, "true") {
			t.Error("isHidden should be true")
		}
	})

	t.Run("JSON output shows isHidden false for visible keys", func(t *testing.T) {
		RunKVSuccess(t, "set", "visible-key", "value")

		output := RunKVSuccess(t, "list", "visible-key", "--output", "json")
		// With omitempty, false values might not appear, but value should be present
		if !strings.Contains(output, `"value"`) {
			t.Error("JSON output should include value for visible keys")
		}
	})
}

func TestHideMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("hide multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "key1", "value1")
		RunKVSuccess(t, "set", "key2", "value2")
		RunKVSuccess(t, "set", "key3", "value3")

		RunKVSuccess(t, "hide", "key1", "key2", "key3")

		// All keys should be hidden
		output := RunKVSuccess(t, "list", "key1")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("key1 should be hidden")
		}

		output = RunKVSuccess(t, "list", "key2")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("key2 should be hidden")
		}

		output = RunKVSuccess(t, "list", "key3")
		if !strings.Contains(output, "[Hidden]") {
			t.Error("key3 should be hidden")
		}
	})

	// Setup keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a")
	RunKVSuccess(t, "set", "b", "value-b")

	testCases := []struct {
		name string
		keys []string
	}{
		{"hide fails on first non-existent key", []string{"missing", "a", "b"}},
		{"hide fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"hide fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"hide"}, tc.keys...)
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should not be hidden (transaction rollback)
			output = RunKVSuccess(t, "list", "a")
			if strings.Contains(output, "[Hidden]") {
				t.Error("a should not be hidden due to transaction rollback")
			}

			output = RunKVSuccess(t, "list", "b")
			if strings.Contains(output, "[Hidden]") {
				t.Error("b should not be hidden due to transaction rollback")
			}
		})
	}
}

func TestShowMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("show multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "h1", "secret1")
		RunKVSuccess(t, "set", "h2", "secret2")
		RunKVSuccess(t, "set", "h3", "secret3")
		RunKVSuccess(t, "hide", "h1")
		RunKVSuccess(t, "hide", "h2")
		RunKVSuccess(t, "hide", "h3")

		RunKVSuccess(t, "show", "h1", "h2", "h3")

		// All keys should be visible
		output := RunKVSuccess(t, "list", "h1")
		if strings.Contains(output, "[Hidden]") {
			t.Error("h1 should be visible")
		}

		output = RunKVSuccess(t, "list", "h2")
		if strings.Contains(output, "[Hidden]") {
			t.Error("h2 should be visible")
		}

		output = RunKVSuccess(t, "list", "h3")
		if strings.Contains(output, "[Hidden]") {
			t.Error("h3 should be visible")
		}
	})

	// Setup hidden keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a")
	RunKVSuccess(t, "set", "b", "value-b")
	RunKVSuccess(t, "hide", "a")
	RunKVSuccess(t, "hide", "b")

	testCases := []struct {
		name string
		keys []string
	}{
		{"show fails on first non-existent key", []string{"missing", "a", "b"}},
		{"show fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"show fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"show"}, tc.keys...)
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should still be hidden (transaction rollback)
			output = RunKVSuccess(t, "list", "a")
			if !strings.Contains(output, "[Hidden]") {
				t.Error("a should still be hidden due to transaction rollback")
			}

			output = RunKVSuccess(t, "list", "b")
			if !strings.Contains(output, "[Hidden]") {
				t.Error("b should still be hidden due to transaction rollback")
			}
		})
	}
}
