package services

import (
	"etop/db"
	"etop/models"
)

func GetTaskStatuses() []models.TaskStatus {
	taskStatuses := []models.TaskStatus{}
	if err := db.PgSql.Find(&taskStatuses).Error; err != nil {
		return []models.TaskStatus{}
	}
	return taskStatuses
}

func GetTaskPriorities() []models.TaskPriority {
	taskPriorities := []models.TaskPriority{}
	if err := db.PgSql.Find(&taskPriorities).Error; err != nil {
		return []models.TaskPriority{}
	}
	return taskPriorities
}

func CountAchievedTasks(userID string) int64 {
	var count int64
	db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Count(&count)
	return count
}

func IsUserWatchingTask(task models.Task, userID string) bool {
	for _, watcher := range task.Watchers {
		if watcher.UserID == userID {
			return true
		}
	}
	return false
}
