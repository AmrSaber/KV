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
