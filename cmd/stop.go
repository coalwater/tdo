package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Hidden: true,
	Use:    "stop",
	Short: "Stop working on a task",
	Long: `Remove the now label from a task, marking it as no longer active.

Examples:
  tdo 3 stop`,
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

		if !task.HasLabel(app.NowLabel) {
			fmt.Fprintf(cmd.OutOrStdout(), "Task '%s' is not started.\n", task.Content)
			return nil
		}

		// Remove now label.
		labels := make([]string, 0, len(task.Labels))
		for _, l := range task.Labels {
			if l != app.NowLabel {
				labels = append(labels, l)
			}
		}

		params := domain.UpdateParams{
			Labels: labels,
		}
		if err := app.Backend.UpdateTask(ctx, result.TaskID, params); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"status": "stopped",
				"id":     result.TaskID,
			})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Stopped task '%s'.\n", task.Content)
		return nil
	},
}

func init() {
	stopCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = stopCmd.Flags().MarkHidden("id")
}
