package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a task",
	Long: `Delete a task permanently.

Examples:
  tdo 3 delete
  tdo abc123 delete`,
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

		if err := app.Backend.DeleteTask(ctx, result.TaskID); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		content := result.TaskID
		if result.Task != nil {
			content = result.Task.Content
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted task '%s'.\n", content)
		return nil
	},
}

func init() {
	deleteCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = deleteCmd.Flags().MarkHidden("id")
}
