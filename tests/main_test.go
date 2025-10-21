package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Build the kv binary before running tests
	buildCmd := exec.Command("go", "build", "-o", "kv")
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build kv binary: " + err.Error())
	}

	// Get absolute path to kv binary
	kvPath, err := filepath.Abs("../kv")
	if err != nil {
		panic("Failed to get absolute path: " + err.Error())
	}

	// Update PATH to include parent directory
	os.Setenv("PATH", filepath.Dir(kvPath)+":"+os.Getenv("PATH"))

	// Run tests
	code := m.Run()

	// Cleanup (must be before os.Exit since defer won't run)
	os.Remove("../kv")

	os.Exit(code)
}
