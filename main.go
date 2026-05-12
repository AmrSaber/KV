package main

import (
	"github.com/AmrSaber/kv/src/cmd"
	"github.com/AmrSaber/kv/src/common"
	"github.com/AmrSaber/kv/src/services"
)

var version string

func main() {
	// Set version number if it's loaded from build
	if version != "" {
		common.SetVersion(version)
	}

	services.MigrateOldDBName()
	services.MigrateInvalidKeys()

	cmd.Execute()
}
