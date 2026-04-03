package domain

import (
	"math"
	"time"
)

// Urgency coefficients matching TaskWarrior defaults.
const (
	urgencyPriorityHCoeff   = 6.0
	urgencyPriorityMCoeff   = 3.9
	urgencyPriorityLCoeff   = 1.8
	urgencyDueCoeff         = 12.0
	urgencyNextCoeff        = 15.0
	urgencyAgeCoeff         = 2.0
	urgencyAnnotationsCoeff = 1.0
	urgencyTagsCoeff        = 1.0
	urgencyProjectCoeff     = 1.0
)

// CalculateUrgency computes a TaskWarrior-compatible urgency score for a task.
// nowLabel is the label that marks a task as active/started.
// now is the reference time for age and due calculations.
func CalculateUrgency(task Task, nowLabel string, now time.Time) float64 {
	var urgency float64

	urgency += priorityFactor(task.Priority)
	urgency += urgencyDueCoeff * dueFactor(task.Due, now)
	urgency += urgencyNextCoeff * activeFactor(task, nowLabel)
	urgency += urgencyAgeCoeff * ageFactor(task.CreatedAt, now)
	urgency += urgencyAnnotationsCoeff * annotationsFactor(task.CommentCount)
	urgency += urgencyTagsCoeff * tagsFactor(len(task.Labels))
	urgency += urgencyProjectCoeff * projectFactor(task.Project)

	return urgency
}

func priorityFactor(p Priority) float64 {
	switch p {
	case PriorityH:
		return urgencyPriorityHCoeff
	case PriorityM:
		return urgencyPriorityMCoeff
	case PriorityL:
		return urgencyPriorityLCoeff
	default:
		return 0.0
	}
}

func dueFactor(due *time.Time, now time.Time) float64 {
	if due == nil {
		return 0.0
	}

	daysOverdue := now.Sub(*due).Seconds() / 86400.0

	switch {
	case daysOverdue >= 7.0:
		return 1.0
	case daysOverdue >= -14.0:
		return ((daysOverdue + 14.0) * 0.8 / 21.0) + 0.2
	default:
		return 0.2
	}
}

func activeFactor(task Task, nowLabel string) float64 {
	if nowLabel != "" && task.HasLabel(nowLabel) {
		return 1.0
	}
	return 0.0
}

func ageFactor(created time.Time, now time.Time) float64 {
	days := now.Sub(created).Hours() / 24.0
	if days < 0 {
		days = 0
	}
	return math.Min(days/365.0, 1.0)
}

func annotationsFactor(count int) float64 {
	switch {
	case count >= 3:
		return 1.0
	case count == 2:
		return 0.9
	case count == 1:
		return 0.8
	default:
		return 0.0
	}
}

func tagsFactor(count int) float64 {
	switch {
	case count >= 3:
		return 1.0
	case count == 2:
		return 0.9
	case count == 1:
		return 0.8
	default:
		return 0.0
	}
}

func projectFactor(project string) float64 {
	if project != "" {
		return 1.0
	}
	return 0.0
}
