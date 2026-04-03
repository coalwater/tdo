package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var jsonOutput bool

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

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
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

// matchCommand resolves a possibly-abbreviated command name against knownCommands.
// Returns the full command name on exact or unambiguous prefix match,
// empty string on no match, or error on ambiguous match.
func matchCommand(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	// Exact match always wins.
	if knownCommands[input] {
		return input, nil
	}

	// Collect prefix matches.
	var matches []string
	for k := range knownCommands {
		if strings.HasPrefix(k, input) {
			matches = append(matches, k)
		}
	}

	switch len(matches) {
	case 0:
		return "", nil
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous command '%s' — matches: %s", input, strings.Join(matches, ", "))
	}
}

// RewriteIDArgs rewrites TaskWarrior-style "tdo <id> <cmd>" into
// "tdo <cmd> --id <id>" so Cobra can route to the correct subcommand.
// Also expands abbreviated command names. Must be called before Execute().
func RewriteIDArgs(args []string) ([]string, error) {
	if len(args) < 2 {
		return args, nil
	}

	// Try to expand args[1] as a command (handles "tdo mod pri:H").
	first := args[1]
	cmd, err := matchCommand(first)
	if err != nil {
		return nil, err
	}
	if cmd != "" {
		args[1] = cmd
		return args, nil
	}

	// args[1] is not a command — treat it as an ID.
	if len(args) < 3 {
		return args, nil
	}

	// Look for a known command (exact or prefix) in args[2:].
	for i := 2; i < len(args); i++ {
		matched, err := matchCommand(args[i])
		if err != nil {
			return nil, err
		}
		if matched != "" {
			// Rewrite: move the command to position 1, inject --id <id>
			newArgs := []string{args[0], matched, "--id", first}
			// Append anything before the command (between first and i) and after it.
			newArgs = append(newArgs, args[2:i]...)
			newArgs = append(newArgs, args[i+1:]...)
			return newArgs, nil
		}
	}

	return args, nil
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
