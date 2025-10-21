package tests

import (
	"strings"
	"testing"
)

func TestLockCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("lock plain text key", func(t *testing.T) {
		RunKVSuccess(t, "set", "plain", "secret")
		RunKVSuccess(t, "lock", "plain", "--password", "mypass")

		// Should show as locked
		output := RunKVSuccess(t, "list", "plain")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Key should be locked")
		}

		// Should decrypt with password
		output = RunKVSuccess(t, "get", "plain", "--password", "mypass")
		if output != "secret" {
			t.Errorf("Expected 'secret', got: %s", output)
		}
	})

	t.Run("lock already locked key fails", func(t *testing.T) {
		RunKVSuccess(t, "set", "locked", "data", "--password", "pass1")
		output := RunKVFailure(t, "lock", "locked", "--password", "pass2")
		if !strings.Contains(output, "already locked") {
			t.Errorf("Expected 'already locked' error, got: %s", output)
		}
	})

	t.Run("lock non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "lock", "non-existent", "--password", "pass")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("lock with prefix", func(t *testing.T) {
		RunKVSuccess(t, "set", "secrets.api", "key1")
		RunKVSuccess(t, "set", "secrets.db", "key2")
		RunKVSuccess(t, "set", "public.data", "data")

		RunKVSuccess(t, "lock", "secrets", "--prefix", "--password", "pass")

		// secrets.* should be locked
		output := RunKVSuccess(t, "list", "secrets")
		if !strings.Contains(output, "[Locked]") {
			t.Error("Keys with prefix should be locked")
		}

		// public.data should not be locked
		output = RunKVSuccess(t, "get", "public.data")
		if output != "data" {
			t.Error("public.data should not be locked")
		}
	})

	t.Run("lock all keys", func(t *testing.T) {
		// Use fresh database to avoid locked keys from previous tests
		cleanup2 := SetupTestDB(t)
		defer cleanup2()

		RunKVSuccess(t, "set", "key1", "value1")
		RunKVSuccess(t, "set", "key2", "value2")

		RunKVSuccess(t, "lock", "--all", "--password", "masterpass")

		// All keys should be locked
		output := RunKVSuccess(t, "list")
		occurrences := strings.Count(output, "[Locked]")
		if occurrences < 2 {
			t.Errorf("Expected at least 2 locked keys, found %d", occurrences)
		}
	})
}

func TestUnlockCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("unlock locked key", func(t *testing.T) {
		RunKVSuccess(t, "set", "encrypted", "data", "--password", "pass")
		RunKVSuccess(t, "unlock", "encrypted", "--password", "pass")

		// Should be plain text now
		output := RunKVSuccess(t, "get", "encrypted")
		if output != "data" {
			t.Errorf("Expected 'data', got: %s", output)
		}

		// Should not show as locked
		output = RunKVSuccess(t, "list", "encrypted")
		if strings.Contains(output, "[Locked]") {
			t.Error("Key should not be locked after unlock")
		}
	})

	t.Run("unlock with wrong password fails", func(t *testing.T) {
		RunKVSuccess(t, "set", "locked", "secret", "--password", "correct")
		output := RunKVFailure(t, "unlock", "locked", "--password", "wrong")
		if !strings.Contains(output, "Wrong password") {
			t.Errorf("Expected 'Wrong password' error, got: %s", output)
		}
	})

	t.Run("unlock plain text key fails", func(t *testing.T) {
		RunKVSuccess(t, "set", "plain", "data")
		output := RunKVFailure(t, "unlock", "plain", "--password", "pass")
		if !strings.Contains(output, "not locked") {
			t.Errorf("Expected 'not locked' error, got: %s", output)
		}
	})

	t.Run("unlock non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "unlock", "non-existent", "--password", "pass")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("unlock with prefix", func(t *testing.T) {
		// Use fresh database to avoid plain text keys from previous tests
		cleanup2 := SetupTestDB(t)
		defer cleanup2()

		RunKVSuccess(t, "set", "enc.key1", "value1", "--password", "pass")
		RunKVSuccess(t, "set", "enc.key2", "value2", "--password", "pass")

		RunKVSuccess(t, "unlock", "enc", "--prefix", "--password", "pass")

		// Both keys should be unlocked
		output := RunKVSuccess(t, "get", "enc.key1")
		if output != "value1" {
			t.Error("enc.key1 should be unlocked")
		}

		output = RunKVSuccess(t, "get", "enc.key2")
		if output != "value2" {
			t.Error("enc.key2 should be unlocked")
		}
	})

	t.Run("unlock all keys", func(t *testing.T) {
		// Use fresh database to avoid plain text keys from previous tests
		cleanup2 := SetupTestDB(t)
		defer cleanup2()

		RunKVSuccess(t, "set", "a", "data1", "--password", "pass")
		RunKVSuccess(t, "set", "b", "data2", "--password", "pass")

		RunKVSuccess(t, "unlock", "--all", "--password", "pass")

		// All keys should be unlocked
		output := RunKVSuccess(t, "list")
		if strings.Contains(output, "[Locked]") {
			t.Error("No keys should be locked after unlock --all")
		}
	})
}
