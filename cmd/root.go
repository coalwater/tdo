package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tdo",
	Short: "TaskWarrior-compatible CLI backed by Todoist",
	Long:  "tdo provides a TaskWarrior-compatible command-line interface using Todoist as the backend. Cloud sync, mobile access, and collaboration — with the CLI you know.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip app init for commands that don't need the backend.
		switch cmd.Name() {
		case "version", "help":
			return nil
		}
		return initApp()
	},
}

// knownCommands is the set of subcommand names for ID-first routing.
var knownCommands = map[string]bool{
	"add":      true,
	"list":     true,
	"ls":       true,
	"next":     true,
	"projects": true,
	"tags":     true,
	"version":  true,
	"help":     true,
	"done":     true,
	"delete":   true,
	"modify":   true,
	"start":    true,
	"stop":     true,
	"info":     true,
	"annotate": true,
}

// RewriteIDArgs rewrites TaskWarrior-style "tdo <id> <cmd>" into
// "tdo <cmd> --id <id>" so Cobra can route to the correct subcommand.
// Must be called before Execute().
func RewriteIDArgs(args []string) []string {
	if len(args) < 3 {
		return args
	}

	first := args[1]
	if knownCommands[first] {
		return args
	}

	// first arg is not a known command — treat it as an ID.
	// Look for a known command in args[2:].
	for i := 2; i < len(args); i++ {
		if knownCommands[args[i]] {
			// Rewrite: move the command to position 1, inject --id <id>
			newArgs := []string{args[0], args[i], "--id", first}
			// Append anything before the command (between first and i) and after it.
			newArgs = append(newArgs, args[2:i]...)
			newArgs = append(newArgs, args[i+1:]...)
			return newArgs
		}
	}

	return args
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(modifyCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(annotateCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(tagsCmd)
}
