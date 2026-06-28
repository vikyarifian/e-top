package services

import (
	"fmt"
	"time"

	"etop/db"
	"etop/models"
)

func AddLog(userID string, action string, resourceType string, resourceID string, details map[string]any) error {
	t := time.Now()
	if err := db.PgSql.Create(&models.Log{UserID: userID, Action: action, ResourceType: resourceType, ResourceID: resourceID, Details: details, CreatedAt: &t, CreatedBy: userID, UpdatedAt: &t, UpdatedBy: userID}).Error; err != nil {
		println(err.Error())
		return err
	}
	return nil
}

func GetUserNotifications(userID string, limit int) []models.Notif {
	var logs []models.Log
	userIDs := []string{userID}
	_ = userIDs

	db.PgSql.Where(`
		user_id = ?
		OR (resource_type = 'Task' AND resource_id IN (SELECT id FROM tasks WHERE (user_id = ? OR created_by = ?)))
		OR (resource_type = 'Project' AND resource_id IN (SELECT project_id FROM project_members WHERE user_id = ?))
	`, userID, userID, userID, userID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)

	notifs := make([]models.Notif, 0, len(logs))
	for _, l := range logs {
		desc := ""
		if l.Details != nil {
			if d, ok := l.Details["description"]; ok {
				desc = fmt.Sprintf("%v", d)
			}
		}
		if desc == "" {
			desc = actionLabel(l.Action)
		}
		actorName := l.User.FullName
		actorColor := l.User.Color
		if actorColor == "" {
			actorColor = "bg-gray-500"
		}

		notifs = append(notifs, models.Notif{
			ID:           l.ID,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceID:   l.ResourceID,
			Description:  desc,
			Timestamp:    *l.CreatedAt,
			ActorName:    actorName,
			ActorColor:   actorColor,
		})
	}
	return notifs
}

func actionLabel(action string) string {
	labels := map[string]string{
		"created_task":    "created a task",
		"updated_task":    "updated a task",
		"completed_task":  "completed a task",
		"added_comment":   "added a comment",
		"created_project": "created a project",
		"updated_project": "updated a project",
		"watched_task":    "started watching a task",
		"unwatched_task":  "stopped watching a task",
	}
	if l, ok := labels[action]; ok {
		return l
	}
	return action
}
