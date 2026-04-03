package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/abushady/tdo/internal/domain"
)

var (
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	valueStyle = lipgloss.NewStyle()
)

const labelWidth = 14

func fieldLine(label, value string) string {
	return fmt.Sprintf("%s  %s\n",
		labelStyle.Render(padRight(label+":", labelWidth)),
		valueStyle.Render(value),
	)
}

// FormatTaskDetail renders a single task with all details.
func FormatTaskDetail(task domain.Task, comments []domain.Comment, urgency float64) string {
	var b strings.Builder

	statusStr := "Pending"
	if task.IsCompleted {
		statusStr = "Completed"
	}
	if task.HasLabel("now") {
		statusStr = "Active"
	}

	b.WriteString(fieldLine("Name", task.Content))
	b.WriteString(fieldLine("ID", task.ID))
	b.WriteString(fieldLine("Status", statusStr))

	if task.Project != "" {
		b.WriteString(fieldLine("Project", task.Project))
	}

	if task.Priority != domain.PriorityNone {
		b.WriteString(fieldLine("Priority", task.Priority.String()))
	}

	if task.Due != nil {
		now := time.Now()
		rel := FormatRelativeDue(*task.Due, now)
		dueStr := fmt.Sprintf("%s (%s)", task.Due.Format("2006-01-02"), rel)
		b.WriteString(fieldLine("Due", dueStr))
	}

	if len(task.Labels) > 0 {
		b.WriteString(fieldLine("Labels", strings.Join(task.Labels, ", ")))
	}

	if task.Description != "" {
		b.WriteString(fieldLine("Description", task.Description))
	}

	if task.Recurrence != "" {
		b.WriteString(fieldLine("Recurrence", task.Recurrence))
	}

	b.WriteString(fieldLine("Urgency", fmt.Sprintf("%.1f", urgency)))
	b.WriteString(fieldLine("Created", task.CreatedAt.Format("2006-01-02")))

	if len(comments) > 0 {
		b.WriteString("\nAnnotations:\n")
		for _, c := range comments {
			fmt.Fprintf(&b, "  %s  %s\n", c.PostedAt.Format("2006-01-02"), c.Content)
		}
	}

	return b.String()
}
