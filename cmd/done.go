package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed.

Examples:
  tdo 3 done
  tdo abc123 done`,
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

		if err := app.Backend.CompleteTask(ctx, result.TaskID); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		content := result.TaskID
		if result.Task != nil {
			content = result.Task.Content
		}

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"status": "done",
				"id":     result.TaskID,
			})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Completed task '%s'.\n", content)
		return nil
	},
}

func init() {
	doneCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = doneCmd.Flags().MarkHidden("id")
}
