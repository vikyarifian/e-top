package features

import (
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
