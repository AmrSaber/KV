package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Helper to run kv command and return output
func RunKV(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Build path to the kv binary
	kvBinary := "../kv"

	cmd := exec.Command(kvBinary, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// Helper to run kv command and expect success
func RunKVSuccess(t *testing.T, args ...string) string {
	t.Helper()
	output, err := RunKV(t, args...)
	if err != nil {
		t.Fatalf("Command failed: kv %v\nError: %v\nOutput: %s", args, err, output)
	}
	return output
}

// Helper to run kv command and expect failure
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
	os.Setenv("XDG_DATA_HOME", tmpDir)

	// Return cleanup function
	return func() {
		os.Unsetenv("XDG_DATA_HOME")
		os.RemoveAll(tmpDir)
	}
}
