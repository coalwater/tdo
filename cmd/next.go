package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/abushady/tdo/internal/display"
	"github.com/abushady/tdo/internal/domain"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next [filter...]",
	Short: "Show most urgent tasks",
	Long: `Show the most urgent tasks, limited to fit the terminal height.

Examples:
  tdo next
  tdo next project:Work +urgent
  tdo next limit:5`,
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

		// Apply limit: explicit limit: filter wins, otherwise cap to terminal height.
		limit := filter.Limit
		if limit == 0 && !jsonOutput {
			maxRows := 25
			if lines := os.Getenv("LINES"); lines != "" {
				if n, err := strconv.Atoi(lines); err == nil && n > 3 {
					maxRows = n
				}
			}
			maxRows -= 3 // header + footer + blank line
			if maxRows < 1 {
				maxRows = 1
			}
			limit = maxRows
		}
		if limit > 0 && limit < len(filtered) {
			filtered = filtered[:limit]
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
