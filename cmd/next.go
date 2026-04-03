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
	Use:   "next",
	Short: "Show most urgent tasks",
	Long:  "Show the most urgent tasks, limited to fit the terminal height.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		now := time.Now()

		tasks, err := app.GetTasks(ctx)
		if err != nil {
			return err
		}

		app.EnrichProjectNames(ctx, tasks)

		// Sort by urgency descending.
		sort.Slice(tasks, func(i, j int) bool {
			ui := domain.CalculateUrgency(tasks[i], app.NowLabel, now)
			uj := domain.CalculateUrgency(tasks[j], app.NowLabel, now)
			return ui > uj
		})

		if !jsonOutput {
			// Limit to terminal height minus header/footer.
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
			if len(tasks) > maxRows {
				tasks = tasks[:maxRows]
			}
		}

		if jsonOutput {
			items := make([]taskJSON, len(tasks))
			positions := make(map[int]string, len(tasks))
			for i, t := range tasks {
				items[i] = toTaskJSON(t, app.NowLabel, now)
				positions[i+1] = t.ID
			}
			_ = app.Cache.SetPositions(positions)
			return writeJSON(cmd.OutOrStdout(), items)
		}

		output, positions := display.FormatTaskTable(tasks, app.NowLabel, now)
		_ = app.Cache.SetPositions(positions)

		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}
