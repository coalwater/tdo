package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/display"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		projects, err := app.GetProjects(ctx)
		if err != nil {
			return err
		}

		output := display.FormatProjectList(projects)
		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}
