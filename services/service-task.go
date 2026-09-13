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
	if err := db.PgSql.Order("no").Find(&taskPriorities).Error; err != nil {
		return []models.TaskPriority{}
	}
	return taskPriorities
}

// GetTaskImpacts mengembalikan acuan luas dampak tugas. Urutannya mengikuti
// kolom no agar pilihan pada formulir selalu tampil dalam urutan yang sama.
func GetTaskImpacts() []models.TaskImpact {
	taskImpacts := []models.TaskImpact{}
	if err := db.PgSql.Order("no").Find(&taskImpacts).Error; err != nil {
		return []models.TaskImpact{}
	}
	return taskImpacts
}

// MaxDueMinutesFor mengembalikan batas tenggat milik sebuah prioritas.
// Nilai 0 berarti prioritas itu tidak dibatasi.
func MaxDueMinutesFor(priorityID int) int {
	for _, p := range GetTaskPriorities() {
		if p.No == priorityID {
			return p.MaxDueMinutes
		}
	}
	return 0
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
