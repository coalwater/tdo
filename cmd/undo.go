package cmd

import (
	"fmt"

	"github.com/abushady/tdo/internal/undo"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last mutating operation",
	Long: `Undo the last mutating operation (done, delete, modify, add, start, stop).

Each undo consumes the most recent entry from the local undo log (up to 10 entries).
The undo log is stored in ~/.cache/tdo/undo_log.json.

Note: undo of delete re-creates the task with a new ID — comments and history are lost.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		entry, err := app.UndoLog.Peek()
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("nothing to undo")
		}

		var description string
		var newID string

		switch entry.Op {
		case undo.OpDone:
			if err := app.Backend.ReopenTask(ctx, entry.TaskID); err != nil {
				return err
			}
			description = "reopened"

		case undo.OpDelete:
			if entry.Snapshot == nil {
				return fmt.Errorf("cannot undo: missing task snapshot")
			}
			params := undo.SnapshotToCreateParams(entry.Snapshot)
			task, err := app.Backend.CreateTask(ctx, params)
			if err != nil {
				return err
			}
			newID = task.ID
			description = "re-created"

		case undo.OpModify, undo.OpStart, undo.OpStop:
			if entry.Snapshot == nil {
				return fmt.Errorf("cannot undo: missing task snapshot")
			}
			params := undo.SnapshotToUpdateParams(entry.Snapshot)
			if err := app.Backend.UpdateTask(ctx, entry.TaskID, params); err != nil {
				return err
			}
			description = "reverted"

		case undo.OpAdd:
			if entry.CreatedID == "" {
				return fmt.Errorf("cannot undo: missing created task ID")
			}
			if err := app.Backend.DeleteTask(ctx, entry.CreatedID); err != nil {
				return err
			}
			description = "removed"

		default:
			return fmt.Errorf("unknown undo op: %s", entry.Op)
		}

		// Backend call succeeded — pop the entry now (atomic: entry stays on failure).
		if _, err := app.UndoLog.Pop(); err != nil {
			return err
		}

		_ = app.Cache.InvalidateTasks()

		content := entry.TaskID
		if entry.Snapshot != nil {
			content = entry.Snapshot.Content
		}
		if entry.Op == undo.OpAdd {
			content = entry.CreatedID
		}

		if jsonOutput {
			out := map[string]string{
				"status":      "undone",
				"op":          string(entry.Op),
				"task_id":     entry.TaskID,
				"description": description,
			}
			if newID != "" {
				out["new_id"] = newID
			}
			return writeJSON(cmd.OutOrStdout(), out)
		}

		switch entry.Op {
		case undo.OpDone:
			fmt.Fprintf(cmd.OutOrStdout(), "Undone: reopened '%s' (was completed)\n", content)
		case undo.OpDelete:
			fmt.Fprintf(cmd.OutOrStdout(), "Undone: re-created '%s' (was deleted, new ID: %s)\n", content, newID)
		case undo.OpModify, undo.OpStart, undo.OpStop:
			fmt.Fprintf(cmd.OutOrStdout(), "Undone: reverted '%s' to previous state\n", content)
		case undo.OpAdd:
			fmt.Fprintf(cmd.OutOrStdout(), "Undone: removed '%s' (was just added)\n", content)
		}

		return nil
	},
}
