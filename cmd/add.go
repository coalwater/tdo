package cmd

import (
	"fmt"
	"time"

	"github.com/abushady/tdo/internal/domain"
	"github.com/abushady/tdo/internal/undo"
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
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		now := time.Now()

		_, jsonOut, help, remaining := extractFlags(args)
		if help {
			return cmd.Help()
		}
		jsonOutput = jsonOut

		if len(remaining) == 0 {
			return fmt.Errorf("accepts 1 arg(s), received 0")
		}

		attrs, err := domain.ParseAttributes(remaining, now)
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

		if app.UndoLog != nil {
			_ = app.UndoLog.Push(undo.Entry{Op: undo.OpAdd, CreatedID: task.ID, Timestamp: time.Now()})
		}
		_ = app.Cache.InvalidateTasks()

		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), toTaskJSON(*task, app.NowLabel, time.Now()))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created task %s.\n", task.ID)
		return nil
	},
}
