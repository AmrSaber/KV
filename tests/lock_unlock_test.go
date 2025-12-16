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

func TestLockMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("lock multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "lk1", "secret1")
		RunKVSuccess(t, "set", "lk2", "secret2")
		RunKVSuccess(t, "set", "lk3", "secret3")

		RunKVSuccess(t, "lock", "lk1", "lk2", "lk3", "--password", "mypass")

		// All keys should be locked
		output := RunKVSuccess(t, "list", "lk1")
		if !strings.Contains(output, "[Locked]") {
			t.Error("lk1 should be locked")
		}

		output = RunKVSuccess(t, "list", "lk2")
		if !strings.Contains(output, "[Locked]") {
			t.Error("lk2 should be locked")
		}

		output = RunKVSuccess(t, "list", "lk3")
		if !strings.Contains(output, "[Locked]") {
			t.Error("lk3 should be locked")
		}

		// Should be able to decrypt with password
		output = RunKVSuccess(t, "get", "lk1", "--password", "mypass")
		if output != "secret1" {
			t.Errorf("Expected 'secret1', got: %s", output)
		}
	})

	// Setup keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a")
	RunKVSuccess(t, "set", "b", "value-b")

	testCases := []struct {
		name string
		keys []string
	}{
		{"lock fails on first non-existent key", []string{"missing", "a", "b"}},
		{"lock fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"lock fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"lock"}, tc.keys...)
			args = append(args, "--password", "pass")
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should not be locked (transaction rollback)
			output = RunKVSuccess(t, "list", "a")
			if strings.Contains(output, "[Locked]") {
				t.Error("a should not be locked due to transaction rollback")
			}

			output = RunKVSuccess(t, "list", "b")
			if strings.Contains(output, "[Locked]") {
				t.Error("b should not be locked due to transaction rollback")
			}
		})
	}

	t.Run("lock fails on already locked key", func(t *testing.T) {
		RunKVSuccess(t, "set", "lk6", "value1")
		RunKVSuccess(t, "set", "lk7", "value2", "--password", "oldpass")
		RunKVSuccess(t, "set", "lk8", "value3")

		// Should fail on second key (already locked)
		output := RunKVFailure(t, "lock", "lk6", "lk7", "lk8", "--password", "newpass")
		if !strings.Contains(output, "already locked") {
			t.Errorf("Expected 'already locked' error, got: %s", output)
		}

		// lk6 should not be locked (transaction rollback)
		output = RunKVSuccess(t, "list", "lk6")
		if strings.Contains(output, "[Locked]") {
			t.Error("lk6 should not be locked due to transaction rollback")
		}

		// lk7 should still be locked with old password
		output = RunKVSuccess(t, "get", "lk7", "--password", "oldpass")
		if output != "value2" {
			t.Error("lk7 should still be locked with old password")
		}

		// lk8 should not be locked
		output = RunKVSuccess(t, "list", "lk8")
		if strings.Contains(output, "[Locked]") {
			t.Error("lk8 should not be locked")
		}
	})
}

func TestUnlockMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("unlock multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "uk1", "secret1", "--password", "pass")
		RunKVSuccess(t, "set", "uk2", "secret2", "--password", "pass")
		RunKVSuccess(t, "set", "uk3", "secret3", "--password", "pass")

		RunKVSuccess(t, "unlock", "uk1", "uk2", "uk3", "--password", "pass")

		// All keys should be unlocked
		output := RunKVSuccess(t, "get", "uk1")
		if output != "secret1" {
			t.Errorf("uk1 should be unlocked, got: %s", output)
		}

		output = RunKVSuccess(t, "get", "uk2")
		if output != "secret2" {
			t.Errorf("uk2 should be unlocked, got: %s", output)
		}

		output = RunKVSuccess(t, "get", "uk3")
		if output != "secret3" {
			t.Errorf("uk3 should be unlocked, got: %s", output)
		}

		// Should not show as locked in list
		output = RunKVSuccess(t, "list", "uk")
		if strings.Contains(output, "[Locked]") {
			t.Error("No keys should show as locked")
		}
	})

	// Setup locked keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a", "--password", "pass")
	RunKVSuccess(t, "set", "b", "value-b", "--password", "pass")

	testCases := []struct {
		name string
		keys []string
	}{
		{"unlock fails on first non-existent key", []string{"missing", "a", "b"}},
		{"unlock fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"unlock fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"unlock"}, tc.keys...)
			args = append(args, "--password", "pass")
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should still be locked (transaction rollback)
			output = RunKVSuccess(t, "list", "a")
			if !strings.Contains(output, "[Locked]") {
				t.Error("a should still be locked due to transaction rollback")
			}

			output = RunKVSuccess(t, "list", "b")
			if !strings.Contains(output, "[Locked]") {
				t.Error("b should still be locked due to transaction rollback")
			}
		})
	}

	t.Run("unlock fails on not locked key", func(t *testing.T) {
		RunKVSuccess(t, "set", "uk6", "value1", "--password", "pass")
		RunKVSuccess(t, "set", "uk7", "value2")
		RunKVSuccess(t, "set", "uk8", "value3", "--password", "pass")

		// Should fail on second key (not locked)
		output := RunKVFailure(t, "unlock", "uk6", "uk7", "uk8", "--password", "pass")
		if !strings.Contains(output, "not locked") {
			t.Errorf("Expected 'not locked' error, got: %s", output)
		}

		// uk6 should still be locked (transaction rollback)
		output = RunKVSuccess(t, "list", "uk6")
		if !strings.Contains(output, "[Locked]") {
			t.Error("uk6 should still be locked due to transaction rollback")
		}

		// uk8 should still be locked
		output = RunKVSuccess(t, "list", "uk8")
		if !strings.Contains(output, "[Locked]") {
			t.Error("uk8 should still be locked")
		}
	})

	t.Run("unlock fails on wrong password", func(t *testing.T) {
		RunKVSuccess(t, "set", "uk9", "value1", "--password", "pass1")
		RunKVSuccess(t, "set", "uk10", "value2", "--password", "pass2")
		RunKVSuccess(t, "set", "uk11", "value3", "--password", "pass1")

		// Should fail on second key (wrong password)
		output := RunKVFailure(t, "unlock", "uk9", "uk10", "uk11", "--password", "pass1")
		if !strings.Contains(output, "Wrong password") {
			t.Errorf("Expected 'Wrong password' error, got: %s", output)
		}

		// All keys should still be locked (transaction rollback)
		output = RunKVSuccess(t, "list", "uk9")
		if !strings.Contains(output, "[Locked]") {
			t.Error("uk9 should still be locked due to transaction rollback")
		}

		output = RunKVSuccess(t, "list", "uk10")
		if !strings.Contains(output, "[Locked]") {
			t.Error("uk10 should still be locked")
		}

		output = RunKVSuccess(t, "list", "uk11")
		if !strings.Contains(output, "[Locked]") {
			t.Error("uk11 should still be locked due to transaction rollback")
		}
	})
}
