package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [attributes...]",
	Short: "Modify a task's attributes",
	Long: `Modify a task's attributes.

Examples:
  tdo 3 modify priority:H due:tomorrow
  tdo 3 modify project:Backend +urgent -old-tag
  tdo 3 modify "New task content"`,
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

		attrs := domain.ParseAttributes(args)

		params := domain.UpdateParams{}

		if attrs.Content != "" {
			params.Content = &attrs.Content
		}
		if attrs.Description != "" {
			params.Description = &attrs.Description
		}
		if attrs.Priority != domain.PriorityNone {
			params.Priority = &attrs.Priority
		}
		if attrs.DueString != "" {
			params.DueString = &attrs.DueString
		}

		// Handle label add/remove by merging with current labels.
		if len(attrs.Labels) > 0 || len(attrs.RemoveLabels) > 0 {
			// Fetch current task to get existing labels.
			task := result.Task
			if task == nil {
				t, err := app.Backend.GetTask(ctx, result.TaskID)
				if err != nil {
					return fmt.Errorf("fetching task for label merge: %w", err)
				}
				task = t
			}

			labelSet := make(map[string]bool, len(task.Labels))
			for _, l := range task.Labels {
				labelSet[l] = true
			}
			for _, l := range attrs.Labels {
				labelSet[l] = true
			}
			for _, l := range attrs.RemoveLabels {
				delete(labelSet, l)
			}

			merged := make([]string, 0, len(labelSet))
			for l := range labelSet {
				merged = append(merged, l)
			}
			params.Labels = merged
		}

		if attrs.Project != "" {
			projectID, err := app.ResolveProjectName(ctx, attrs.Project)
			if err != nil {
				return err
			}
			params.ProjectID = &projectID
		}

		if err := app.Backend.UpdateTask(ctx, result.TaskID, params); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		content := result.TaskID
		if result.Task != nil {
			content = result.Task.Content
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Modified task '%s'.\n", content)
		return nil
	},
}

func init() {
	modifyCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = modifyCmd.Flags().MarkHidden("id")
}
