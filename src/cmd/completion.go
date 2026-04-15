package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AmrSaber/kv/src/common"
	"github.com/spf13/cobra"
)

var SupportedShells = []string{"bash", "zsh", "fish", "powershell"}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: ` # Auto-detect shell (recommended)
  eval "$(kv completion)"

  # Explicit shell:
  eval "$(kv completion bash)"
  eval "$(kv completion zsh)"
  eval "$(kv completion fish)"
  eval "$(kv completion powershell)"`,
	DisableFlagsInUseLine: true,
	ValidArgs:             SupportedShells,
	Args:                  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		var shell string
		if len(args) == 1 {
			shell = args[0]
		} else {
			shell = filepath.Base(os.Getenv("SHELL"))

			if shell == "." {
				shell = ""
			}
		}

		switch shell {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			if shell == "" {
				common.Fail("error: could not detect shell, please pass it explicitly: kv completion <shell>")
			} else {
				common.Fail("error: unsupported shell %q, supported: %v", shell, strings.Join(SupportedShells, ", "))
			}
		}

		common.FailOn(err)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
