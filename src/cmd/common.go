package cmd

import (
	"database/sql"

	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

func completeKeyArg(toComplete string, matchType services.MatchType) ([]cobra.Completion, cobra.ShellCompDirective) {
	var matchingKeys []string
	services.RunInTransaction(func(tx *sql.Tx) {
		matchingKeys = services.ListKeys(tx, toComplete, matchType)
	})

	return []cobra.Completion(matchingKeys), cobra.ShellCompDirectiveNoFileComp
}
