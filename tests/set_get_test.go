package tests

import (
	"strings"
	"testing"
)

func TestSetCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("basic set", func(t *testing.T) {
		RunKVSuccess(t, "set", "key1", "value1")
		output := RunKVSuccess(t, "get", "key1")
		if output != "value1" {
			t.Errorf("Expected 'value1', got: %s", output)
		}
	})

	t.Run("set with spaces", func(t *testing.T) {
		RunKVSuccess(t, "set", "my-key", "value with spaces")
		output := RunKVSuccess(t, "get", "my-key")
		if output != "value with spaces" {
			t.Errorf("Expected 'value with spaces', got: %s", output)
		}
	})

	t.Run("set with encryption", func(t *testing.T) {
		RunKVSuccess(t, "set", "secret", "sensitive data", "--password", "mypass")
		output := RunKVSuccess(t, "get", "secret", "--password", "mypass")
		if output != "sensitive data" {
			t.Errorf("Expected 'sensitive data', got: %s", output)
		}
	})

	t.Run("set with TTL", func(t *testing.T) {
		RunKVSuccess(t, "set", "temp", "temporary", "--expires-after", "1h")
		output := RunKVSuccess(t, "ttl", "temp")
		if !strings.Contains(output, "expires at") {
			t.Errorf("Expected TTL to be set, got: %s", output)
		}
	})

	t.Run("update existing key", func(t *testing.T) {
		RunKVSuccess(t, "set", "update-key", "old value")
		RunKVSuccess(t, "set", "update-key", "new value")
		output := RunKVSuccess(t, "get", "update-key")
		if output != "new value" {
			t.Errorf("Expected 'new value', got: %s", output)
		}
	})

	t.Run("set empty value fails", func(t *testing.T) {
		output := RunKVFailure(t, "set", "empty-key", "")
		if !strings.Contains(output, "No value") && !strings.Contains(output, "value") {
			t.Errorf("Expected error about no value, got: %s", output)
		}
	})
}

func TestGetCommand(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("get existing key", func(t *testing.T) {
		RunKVSuccess(t, "set", "test-key", "test-value")
		output := RunKVSuccess(t, "get", "test-key")
		if output != "test-value" {
			t.Errorf("Expected 'test-value', got: %s", output)
		}
	})

	t.Run("get non-existent key fails", func(t *testing.T) {
		output := RunKVFailure(t, "get", "non-existent")
		if !strings.Contains(output, "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %s", output)
		}
	})

	t.Run("get encrypted key with wrong password fails", func(t *testing.T) {
		RunKVSuccess(t, "set", "encrypted", "secret", "--password", "correct")
		output := RunKVFailure(t, "get", "encrypted", "--password", "wrong")
		if !strings.Contains(output, "Wrong password") {
			t.Errorf("Expected 'Wrong password' error, got: %s", output)
		}
	})

	t.Run("get encrypted key without password fails", func(t *testing.T) {
		RunKVSuccess(t, "set", "locked", "data", "--password", "pass")
		output := RunKVFailure(t, "get", "locked")
		if !strings.Contains(output, "locked") && !strings.Contains(output, "password") {
			t.Errorf("Expected locked/password error, got: %s", output)
		}
	})
}
