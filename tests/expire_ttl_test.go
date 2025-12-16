package tests

import (
	"strings"
	"testing"
	"time"
)

func TestExpireCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("set expiration on existing key", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data")
		RunKVSuccess(t, "expire", "temp", "--after", "1h")

		output := RunKVSuccess(t, "ttl", "temp")
		if !strings.Contains(output, "expires at") {
			t.Errorf("Key should have expiration, got: %s", output)
		}
	})

	t.Run("remove expiration", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "1h")
		RunKVSuccess(t, "expire", "temp", "--never")

		output, _ := RunKV(t, "ttl", "temp")
		if !strings.Contains(output, "does not expire") {
			t.Errorf("Key should not expire, got: %s", output)
		}
	})

	t.Run("expire non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "expire", "non-existent", "--after", "1h")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("update expiration time", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "1h")
		RunKVSuccess(t, "expire", "temp", "--after", "2h")

		output := RunKVSuccess(t, "ttl", "temp")
		if !strings.Contains(output, "expires at") {
			t.Error("Key should still have expiration")
		}
	})

	t.Run("keys with ttl in the past expire", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "-1h")

		output := RunKVFailure(t, "get", "temp")
		if !strings.Contains(output, "does not exist") {
			t.Error("Key should have expired")
		}
	})

	t.Run("keys with ttl actually expires", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "1s")

		time.Sleep(2 * time.Second)

		output := RunKVFailure(t, "get", "temp")
		if !strings.Contains(output, "does not exist") {
			t.Error("Key should have expired")
		}
	})
}

func TestTTLCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("check TTL of key with expiration", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "1h")
		output := RunKVSuccess(t, "ttl", "temp")

		if !strings.Contains(output, "expires at") {
			t.Errorf("Should show expiration time, got: %s", output)
		}
	})

	t.Run("check TTL of key without expiration", func(t *testing.T) {
		RunKVSuccess(t, "set", "permanent", "data")
		output := RunKVFailure(t, "ttl", "permanent")

		if !strings.Contains(output, "does not expire") {
			t.Errorf("Should indicate no expiration, got: %s", output)
		}
	})

	t.Run("check TTL with date flag", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "1h")
		output := RunKVSuccess(t, "ttl", "temp", "--date")

		// Should show just the date, not the countdown
		if strings.Contains(output, "expires at") {
			t.Error("With --date flag should only show timestamp")
		}
	})

	t.Run("TTL of non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "ttl", "non-existent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("TTL shows countdown", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "data", "--expires-after", "30m")
		output := RunKVSuccess(t, "ttl", "temp")

		// Should show time in minutes or similar
		if !strings.Contains(output, "m") && !strings.Contains(output, "s") {
			t.Errorf("Should show time unit (m/s), got: %s", output)
		}
	})
}

func TestExpireMultipleKeys(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("expire multiple keys successfully", func(t *testing.T) {
		RunKVSuccess(t, "set", "exp1", "value1")
		RunKVSuccess(t, "set", "exp2", "value2")
		RunKVSuccess(t, "set", "exp3", "value3")

		RunKVSuccess(t, "expire", "exp1", "exp2", "exp3", "--after", "1h")

		// All keys should have expiration
		output := RunKVSuccess(t, "ttl", "exp1")
		if !strings.Contains(output, "expires at") {
			t.Error("exp1 should have expiration")
		}

		output = RunKVSuccess(t, "ttl", "exp2")
		if !strings.Contains(output, "expires at") {
			t.Error("exp2 should have expiration")
		}

		output = RunKVSuccess(t, "ttl", "exp3")
		if !strings.Contains(output, "expires at") {
			t.Error("exp3 should have expiration")
		}
	})

	t.Run("remove expiration from multiple keys", func(t *testing.T) {
		RunKVSuccess(t, "set", "nexp1", "value1", "--expires-after", "1h")
		RunKVSuccess(t, "set", "nexp2", "value2", "--expires-after", "1h")
		RunKVSuccess(t, "set", "nexp3", "value3", "--expires-after", "1h")

		RunKVSuccess(t, "expire", "nexp1", "nexp2", "nexp3", "--never")

		// All keys should not have expiration
		output, _ := RunKV(t, "ttl", "nexp1")
		if !strings.Contains(output, "does not expire") {
			t.Error("nexp1 should not have expiration")
		}

		output, _ = RunKV(t, "ttl", "nexp2")
		if !strings.Contains(output, "does not expire") {
			t.Error("nexp2 should not have expiration")
		}

		output, _ = RunKV(t, "ttl", "nexp3")
		if !strings.Contains(output, "does not expire") {
			t.Error("nexp3 should not have expiration")
		}
	})

	// Setup keys for transaction rollback tests
	RunKVSuccess(t, "set", "a", "value-a")
	RunKVSuccess(t, "set", "b", "value-b")

	testCases := []struct {
		name string
		keys []string
	}{
		{"expire fails on first non-existent key", []string{"missing", "a", "b"}},
		{"expire fails on middle non-existent key", []string{"a", "missing", "b"}},
		{"expire fails on last non-existent key", []string{"a", "b", "missing"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"expire"}, tc.keys...)
			args = append(args, "--after", "1h")
			output := RunKVFailure(t, args...)
			if !strings.Contains(output, "does not exist") {
				t.Errorf("Expected 'does not exist' error, got: %s", output)
			}

			// Keys should not have expiration (transaction rollback)
			output, _ = RunKV(t, "ttl", "a")
			if !strings.Contains(output, "does not expire") {
				t.Error("a should not have expiration due to transaction rollback")
			}

			output, _ = RunKV(t, "ttl", "b")
			if !strings.Contains(output, "does not expire") {
				t.Error("b should not have expiration due to transaction rollback")
			}
		})
	}

	t.Run("expire multiple keys with --never after partial failure", func(t *testing.T) {
		RunKVSuccess(t, "set", "nexp4", "value1", "--expires-after", "1h")
		RunKVSuccess(t, "set", "nexp5", "value2", "--expires-after", "1h")

		// First try with missing key - should fail
		output := RunKVFailure(t, "expire", "nexp4", "missing", "nexp5", "--never")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}

		// Keys should still have expiration (transaction rollback)
		output = RunKVSuccess(t, "ttl", "nexp4")
		if !strings.Contains(output, "expires at") {
			t.Error("nexp4 should still have expiration due to transaction rollback")
		}

		output = RunKVSuccess(t, "ttl", "nexp5")
		if !strings.Contains(output, "expires at") {
			t.Error("nexp5 should still have expiration due to transaction rollback")
		}

		// Now expire them successfully
		RunKVSuccess(t, "expire", "nexp4", "nexp5", "--never")

		output, _ = RunKV(t, "ttl", "nexp4")
		if !strings.Contains(output, "does not expire") {
			t.Error("nexp4 should not have expiration after successful expire")
		}

		output, _ = RunKV(t, "ttl", "nexp5")
		if !strings.Contains(output, "does not expire") {
			t.Error("nexp5 should not have expiration after successful expire")
		}
	})
}
