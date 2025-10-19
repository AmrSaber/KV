package cmd

import (
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

func completeKeyArg(toComplete string, matchType services.MatchType) ([]cobra.Completion, cobra.ShellCompDirective) {
	matchingKeys := services.ListKeys(toComplete, matchType)
	return []cobra.Completion(matchingKeys), cobra.ShellCompDirectiveNoFileComp
}
