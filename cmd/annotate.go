package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var annotateCmd = &cobra.Command{
	Hidden: true,
	Use:    "annotate <text>",
	Short:  "Add a comment to a task",
	Long: `Add a comment/annotation to a task.

Examples:
  tdo 3 annotate "Found the root cause"
  tdo abc123 annotate Blocked by upstream API`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			return fmt.Errorf("task ID is required")
		}

		result, err := app.ResolveTaskID(ctx, id)
		if err != nil {
			return err
		}

		text := strings.Join(args, " ")

		if _, err := app.Backend.AddComment(ctx, result.TaskID, text); err != nil {
			return err
		}

		content := result.TaskID
		if result.Task != nil {
			content = result.Task.Content
		}

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"status": "annotated",
				"id":     result.TaskID,
			})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Annotated task '%s'.\n", content)
		return nil
	},
}

func init() {
	annotateCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = annotateCmd.Flags().MarkHidden("id")
}
