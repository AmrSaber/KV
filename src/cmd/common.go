package cmd

import (
	"github.com/AmrSaber/kv/src/services"
	"github.com/spf13/cobra"
)

func completeKeyArg(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	matchingKeys := services.MatchExistingKeysByPrefix(nil, toComplete)
	return []cobra.Completion(matchingKeys), cobra.ShellCompDirectiveNoFileComp
}
