package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/display"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all labels/tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		labels, err := app.GetLabels(ctx)
		if err != nil {
			return err
		}

		if jsonOutput {
			items := make([]labelJSON, len(labels))
			for i, l := range labels {
				items[i] = labelJSON{ID: l.ID, Name: l.Name}
			}
			return writeJSON(cmd.OutOrStdout(), items)
		}

		output := display.FormatLabelList(labels)
		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}
