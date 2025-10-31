package common

import (
	"bytes"
	"os"
	"os/exec"
)

var version string

// SetVersion sets the version string (typically injected by goreleaser at build time)
func SetVersion(v string) {
	version = v
}

// GetVersion returns the version string. If version was set via SetVersion (goreleaser build),
// it returns that. Otherwise, it attempts to get the version from git tags.
func GetVersion() string {
	if version != "" {
		return version
	}

	gitVersion := getVersionFromGit()
	if gitVersion != "" {
		return gitVersion
	}

	return ""
}

func getVersionFromGit() string {
	// Check if we're in a git repository
	if _, err := os.Stat(".git"); err != nil {
		return ""
	}

	// Get the latest git tag
	tag, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		return ""
	}

	tag = bytes.TrimSpace(tag)

	// Append '+' if repo has uncommitted changes
	if isRepoDirty() {
		tag = append(tag, '+')
	}

	return string(tag)
}

func isRepoDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, _ := cmd.Output()
	output = bytes.TrimSpace(output)

	// Empty output = Clean repo
	return len(output) != 0
}
