package main

import (
	"github.com/AmrSaber/kv/src/cmd"
	"github.com/AmrSaber/kv/src/common"
)

var version string

func main() {
	// Set version number if it's loaded from build
	if version != "" {
		common.SetVersion(version)
	}

	cmd.Execute()
}
