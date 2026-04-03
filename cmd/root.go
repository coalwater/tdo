package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tdo",
	Short: "TaskWarrior-compatible CLI backed by Todoist",
	Long:  "tdo provides a TaskWarrior-compatible command-line interface using Todoist as the backend. Cloud sync, mobile access, and collaboration — with the CLI you know.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
