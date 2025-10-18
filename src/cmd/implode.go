package cmd

import (
	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

// implodeCmd represents the implode command
var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Clear all data (leaving any configurations)",
	Run: func(cmd *cobra.Command, args []string) {
		common.ClearDB()
	},
}

func init() {
	rootCmd.AddCommand(implodeCmd)
}
