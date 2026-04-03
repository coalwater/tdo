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

		output := display.FormatLabelList(labels)
		fmt.Fprint(cmd.OutOrStdout(), output)
		return nil
	},
}
