package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for tdo.

To load completions:

Bash:

  $ source <(tdo completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ tdo completion bash > /etc/bash_completion.d/tdo
  # macOS:
  $ tdo completion bash > $(brew --prefix)/etc/bash_completion.d/tdo

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ tdo completion zsh > "${fpath[1]}/_tdo"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ tdo completion fish | source

  # To load completions for each session, execute once:
  $ tdo completion fish > ~/.config/fish/completions/tdo.fish

PowerShell:

  PS> tdo completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> tdo completion powershell > tdo.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

// getProjectNames returns a list of project names for shell completion.
func getProjectNames(ctx context.Context) ([]string, error) {
	projects, err := app.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Name
	}
	return names, nil
}
