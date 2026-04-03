package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/abushady/tdo/internal/display"
	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [filter...]",
	Short: "List tasks",
	Long: `List tasks with optional filters.

Examples:
  tdo list
  tdo list project:Backend priority:H +urgent
  tdo list due.before:2024-12-31`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		now := time.Now()

		tasks, err := app.GetTasks(ctx)
		if err != nil {
			return err
		}

		app.EnrichProjectNames(ctx, tasks)

		// Apply filters.
		filter, err := domain.ParseFilter(args)
		if err != nil {
			return err
		}
		var filtered []domain.Task
		for _, t := range tasks {
			if filter.Match(t) {
				filtered = append(filtered, t)
			}
		}

		// Sort by urgency descending.
		sort.Slice(filtered, func(i, j int) bool {
			ui := domain.CalculateUrgency(filtered[i], app.NowLabel, now)
			uj := domain.CalculateUrgency(filtered[j], app.NowLabel, now)
			return ui > uj
		})

		// Apply limit after sorting.
		if filter.Limit > 0 && filter.Limit < len(filtered) {
			filtered = filtered[:filter.Limit]
		}

		if jsonOutput {
			items := make([]taskJSON, len(filtered))
			positions := make(map[int]string, len(filtered))
			for i, t := range filtered {
				items[i] = toTaskJSON(t, app.NowLabel, now)
				positions[i+1] = t.ID
			}
			_ = app.Cache.SetPositions(positions)
			return writeJSON(cmd.OutOrStdout(), items)
		}

		output, positions := display.FormatTaskTable(filtered, app.NowLabel, now)
		_ = app.Cache.SetPositions(positions)

		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}
