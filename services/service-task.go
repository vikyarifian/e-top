package services

import (
	"etop/db"
	"etop/models"
)

func GetTaskStatuses() []models.TaskStatus {
	ts := []models.TaskStatus{}
	db.PgSql.Find(&ts)
	return ts
}

func GetTaskPriorities() []models.TaskPriority {
	prts := []models.TaskPriority{}
	db.PgSql.Find(&prts)
	return prts
}
