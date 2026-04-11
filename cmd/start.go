package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Hidden: true,
	Use:    "start",
	Short: "Start working on a task",
	Long: `Mark a task as actively being worked on by adding the now label.

Examples:
  tdo 3 start`,
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

		// Fetch current task to get existing labels.
		task := result.Task
		if task == nil {
			t, err := app.Backend.GetTask(ctx, result.TaskID)
			if err != nil {
				return err
			}
			task = t
		}

		// Add now label if not already present.
		if task.HasLabel(app.NowLabel) {
			fmt.Fprintf(cmd.OutOrStdout(), "Task '%s' is already started.\n", task.Content)
			return nil
		}

		labels := make([]string, len(task.Labels), len(task.Labels)+1)
		copy(labels, task.Labels)
		labels = append(labels, app.NowLabel)

		params := domain.UpdateParams{
			Labels: labels,
		}
		if err := app.Backend.UpdateTask(ctx, result.TaskID, params); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"status": "started",
				"id":     result.TaskID,
			})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Started task '%s'.\n", task.Content)
		return nil
	},
}

func init() {
	startCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = startCmd.Flags().MarkHidden("id")
}
