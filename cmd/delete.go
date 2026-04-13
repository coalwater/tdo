package cmd

import (
	"fmt"
	"time"

	"github.com/abushady/tdo/internal/undo"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Hidden: true,
	Use:    "delete",
	Short:  "Delete a task",
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

		// Always fetch full task for delete: we need all fields for re-creation.
		task, _ := app.Backend.GetTask(ctx, result.TaskID)
		if task == nil {
			task = result.Task
		}
		if app.UndoLog != nil {
			_ = app.UndoLog.Push(undo.Entry{Op: undo.OpDelete, TaskID: result.TaskID, Snapshot: task, Timestamp: time.Now()})
		}

		if err := app.Backend.DeleteTask(ctx, result.TaskID); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		content := result.TaskID
		if result.Task != nil {
			content = result.Task.Content
		}

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"status": "deleted",
				"id":     result.TaskID,
			})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Deleted task '%s'.\n", content)
		return nil
	},
}

func init() {
	deleteCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = deleteCmd.Flags().MarkHidden("id")
}
