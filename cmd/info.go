package cmd

import (
	"fmt"
	"time"

	"github.com/abushady/tdo/internal/display"
	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show detailed task information",
	Long: `Show detailed information about a task including comments.

Examples:
  tdo 3 info
  tdo abc123 info`,
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

		// Always fetch fresh task details for info.
		task, err := app.Backend.GetTask(ctx, result.TaskID)
		if err != nil {
			return err
		}

		// Enrich project name.
		if task.ProjectID != "" {
			projects, err := app.GetProjects(ctx)
			if err == nil {
				for _, p := range projects {
					if p.ID == task.ProjectID {
						task.Project = p.Name
						break
					}
				}
			}
		}

		comments, err := app.Backend.ListComments(ctx, result.TaskID)
		if err != nil {
			return fmt.Errorf("fetching comments: %w", err)
		}

		now := time.Now()

		if jsonOutput {
			cj := make([]commentJSON, len(comments))
			for i, c := range comments {
				cj[i] = toCommentJSON(c)
			}
			return writeJSON(cmd.OutOrStdout(), map[string]any{
				"task":     toTaskJSON(*task, app.NowLabel, now),
				"comments": cj,
			})
		}

		urgency := domain.CalculateUrgency(*task, app.NowLabel, now)
		output := display.FormatTaskDetail(*task, comments, urgency)

		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}

func init() {
	infoCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
	_ = infoCmd.Flags().MarkHidden("id")
}
