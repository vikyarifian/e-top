package features

import (
	"fmt"
	"time"

	"etop/dto"
	"etop/models"
)

func canEditTask(task models.Task, user dto.UserAuth) bool {
	if task.CreatedBy == user.ID || user.Level == "ADMIN" {
		return true
	}
	for _, member := range task.Project.Members {
		if member.UserID == user.ID && (member.Role == "ADMIN" || member.Role == "OWNER") {
			return true
		}
	}
	return false
}

func taskDueDate(t *time.Time) string {
	if t != nil {
		return t.Format("2006-01-02")
	}
	return ""
}

func taskAssigneeList(task models.Task) []models.UserRole {
	if task.Assignee != nil {
		return []models.UserRole{{ID: task.Assignee.ID, FullName: task.Assignee.FullName, Color: task.Assignee.Color}}
	}
	return []models.UserRole{}
}

func tern(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func toggleSortDir(field string, page models.PageInfo) string {
	if page.SortBy == field {
		if page.SortDir == "asc" {
			return "desc"
		}
		return "asc"
	}
	return "asc"
}

func clampEval(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func evalBadgeColor(cat string) string {
	switch cat {
	case "Sangat Baik":
		return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
	case "Baik":
		return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
	case "Cukup":
		return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200"
	case "Buruk":
		return "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200"
	case "Sangat Buruk":
		return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
	default:
		return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200"
	}
}

func evalScoreColor(cat string) string {
	switch cat {
	case "Sangat Baik":
		return "text-green-600 dark:text-green-400"
	case "Baik":
		return "text-blue-600 dark:text-blue-400"
	case "Cukup":
		return "text-yellow-600 dark:text-yellow-400"
	case "Buruk":
		return "text-orange-600 dark:text-orange-400"
	case "Sangat Buruk":
		return "text-red-600 dark:text-red-400"
	default:
		return "text-gray-600 dark:text-gray-400"
	}
}

func evalProgressColor(cat string) string {
	switch cat {
	case "Sangat Baik":
		return "bg-green-500"
	case "Baik":
		return "bg-blue-500"
	case "Cukup":
		return "bg-yellow-500"
	case "Buruk":
		return "bg-orange-500"
	case "Sangat Buruk":
		return "bg-red-500"
	default:
		return "bg-gray-500"
	}
}

func achieveURL(base string, page int, sortBy, sortDir, year, userID string) string {
	url := fmt.Sprintf("%s?page=%d&sort_by=%s&sort_dir=%s", base, page, sortBy, sortDir)
	if year != "" {
		url += "&year=" + year
	}
	if userID != "" {
		url += "&user_id=" + userID
	}
	return url
}
