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

Attributes:
  due:<expr>        Hard deadline (parsed date expression, e.g. friday, eom-1d, 2026-05-01)
  scheduled:<value> When to work on it (Todoist NLP, e.g. tomorrow, next monday)

Examples:
  tdo add "Fix login bug" project:Backend priority:H due:friday +urgent
  tdo add "Buy groceries" due:eom scheduled:tomorrow +shopping`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		now := time.Now()
		attrs, err := domain.ParseAttributes(args, now)
		if err != nil {
			return err
		}

		params := domain.CreateParams{
			Content:         attrs.Content,
			Description:     attrs.Description,
			Priority:        attrs.Priority,
			ScheduledString: attrs.ScheduledString,
			DueDate:         attrs.DueDate,
			Labels:          attrs.Labels,
			Recurrence:      attrs.Recurrence,
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
