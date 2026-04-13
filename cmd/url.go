package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var urlCmd = &cobra.Command{
	Hidden: true,
	Use:    "url",
	Short:  "Print the Todoist web URL for a task",
	Long: `Print the Todoist web URL for a task.

Examples:
  tdo 3 url
  tdo abc123 url`,
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

		task := result.Task
		if task == nil {
			t, err := app.Backend.GetTask(ctx, result.TaskID)
			if err != nil {
				return err
			}
			task = t
		}

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), map[string]string{
				"id":  task.ID,
				"url": task.URL,
			})
		}

		fmt.Fprintln(cmd.OutOrStdout(), task.URL)
		return nil
	},
}

func init() {
	urlCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = urlCmd.Flags().MarkHidden("id")
}
