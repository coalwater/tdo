package cmd

import (
	"fmt"
	"time"

	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <content> [attributes...]",
	Short: "Create a new task",
	Long: `Create a new task with optional attributes.

Examples:
  tdo add "Fix login bug" project:Backend priority:H due:tomorrow +urgent
  tdo add "Buy groceries" due:today +shopping`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		attrs := domain.ParseAttributes(args)

		params := domain.CreateParams{
			Content:     attrs.Content,
			Description: attrs.Description,
			Priority:    attrs.Priority,
			DueString:   attrs.DueString,
			Labels:      attrs.Labels,
			Recurrence:  attrs.Recurrence,
		}

		if attrs.Project != "" {
			projectID, err := app.ResolveProjectName(ctx, attrs.Project)
			if err != nil {
				return err
			}
			params.ProjectID = projectID
		}

		if attrs.ParentID != "" {
			result, err := app.ResolveTaskID(ctx, attrs.ParentID)
			if err != nil {
				return fmt.Errorf("resolving parent task: %w", err)
			}
			params.ParentID = result.TaskID
		}

		task, err := app.Backend.CreateTask(ctx, params)
		if err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), toTaskJSON(*task, app.NowLabel, time.Now()))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created task %s.\n", task.ID)
		return nil
	},
}
