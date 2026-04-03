package display

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/abushady/tdo/internal/domain"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	normalStyle   = lipgloss.NewStyle()
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	overdueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	dueTodayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
)

const (
	colNum     = 3
	colID      = 18
	colContent = 35
	colPri     = 4
	colLabels  = 12
	colDue     = 12
	colProject = 12
	colUrg     = 7
)

// FormatRelativeDue formats a due date relative to now.
func FormatRelativeDue(due time.Time, now time.Time) string {
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := int(math.Round(dueDay.Sub(nowDay).Hours() / 24))

	switch days {
	case 0:
		return "today"
	case 1:
		return "tomorrow"
	case -1:
		return "yesterday"
	default:
		if days > 14 {
			return due.Format("2006-01-02")
		}
		if days > 0 {
			return fmt.Sprintf("+%dd", days)
		}
		return fmt.Sprintf("%dd", days)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func formatLabels(labels []string) string {
	return truncate(strings.Join(labels, ","), colLabels)
}

func priorityStr(p domain.Priority) string {
	return p.String()
}

type rowStatus int

const (
	statusNormal rowStatus = iota
	statusActive
	statusOverdue
	statusDueToday
)

func classifyTask(t domain.Task, nowLabel string, now time.Time) rowStatus {
	if nowLabel != "" && t.HasLabel(nowLabel) {
		return statusActive
	}
	if t.Due != nil {
		dueDay := time.Date(t.Due.Year(), t.Due.Month(), t.Due.Day(), 0, 0, 0, 0, time.UTC)
		nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if dueDay.Before(nowDay) {
			return statusOverdue
		}
		if dueDay.Equal(nowDay) {
			return statusDueToday
		}
	}
	return statusNormal
}

func styleForStatus(s rowStatus) lipgloss.Style {
	switch s {
	case statusActive:
		return activeStyle
	case statusOverdue:
		return overdueStyle
	case statusDueToday:
		return dueTodayStyle
	default:
		return normalStyle
	}
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-len(s)) + s
}

// FormatTaskTable renders a list of tasks as a colored table.
// Returns the formatted string and a position map (display # -> task ID).
func FormatTaskTable(tasks []domain.Task, nowLabel string, now time.Time) (string, map[int]string) {
	posMap := make(map[int]string, len(tasks))
	if len(tasks) == 0 {
		return "No tasks found.\n", posMap
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("%s  %s  %s  %s  %s  %s  %s  %s",
		padRight("#", colNum),
		padRight("ID", colID),
		padRight("Content", colContent),
		padRight("Pri", colPri),
		padRight("Labels", colLabels),
		padRight("Due", colDue),
		padRight("Project", colProject),
		padLeft("Urg", colUrg),
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	for i, t := range tasks {
		pos := i + 1
		posMap[pos] = t.ID

		dueStr := ""
		if t.Due != nil {
			dueStr = FormatRelativeDue(*t.Due, now)
		}

		urgency := domain.CalculateUrgency(t, nowLabel, now)

		row := fmt.Sprintf("%s  %s  %s  %s  %s  %s  %s  %s",
			padRight(fmt.Sprintf("%d", pos), colNum),
			padRight(truncate(t.ID, colID), colID),
			padRight(truncate(t.Content, colContent), colContent),
			padRight(priorityStr(t.Priority), colPri),
			padRight(formatLabels(t.Labels), colLabels),
			padRight(dueStr, colDue),
			padRight(truncate(t.Project, colProject), colProject),
			padLeft(fmt.Sprintf("%.1f", urgency), colUrg),
		)

		status := classifyTask(t, nowLabel, now)
		style := styleForStatus(status)
		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	return b.String(), posMap
}

// FormatProjectList renders projects as a simple list.
func FormatProjectList(projects []domain.Project) string {
	if len(projects) == 0 {
		return "No projects found.\n"
	}
	var b strings.Builder
	for _, p := range projects {
		fmt.Fprintf(&b, "  %s  %s\n", p.ID, p.Name)
	}
	return b.String()
}

// FormatLabelList renders labels as a simple list.
func FormatLabelList(labels []domain.Label) string {
	if len(labels) == 0 {
		return "No labels found.\n"
	}
	var b strings.Builder
	for _, l := range labels {
		fmt.Fprintf(&b, "  %s  %s\n", l.ID, l.Name)
	}
	return b.String()
}
