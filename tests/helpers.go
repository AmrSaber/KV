// Package tests contains integration tests for the tool
package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var BinaryLocation string

func RunKV(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(BinaryLocation, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func RunKVSuccess(t *testing.T, args ...string) string {
	t.Helper()
	output, err := RunKV(t, args...)
	if err != nil {
		t.Fatalf("Command failed: kv %v\nError: %v\nOutput: %s", args, err, output)
	}
	return output
}

func RunKVFailure(t *testing.T, args ...string) string {
	t.Helper()
	output, err := RunKV(t, args...)
	if err == nil {
		t.Fatalf("Command should have failed: kv %v\nOutput: %s", args, output)
	}
	return output
}

// SetupTestDB creates a temporary database for testing
func SetupTestDB(t *testing.T) func() {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "kv-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Set XDG_DATA_HOME to use temp directory
	// This affects go-application-paths package and it changes data location
	_ = os.Setenv("XDG_DATA_HOME", tmpDir)

	// Return cleanup function
	return func() {
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.RemoveAll(tmpDir)
	}
}
