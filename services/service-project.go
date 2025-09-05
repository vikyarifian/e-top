package services

import (
	"etop/db"
	"etop/models"
)

func GetProjectStatus() []models.ProjectStatus {
	ps := []models.ProjectStatus{}
	db.PgSql.Find(&ps)

	return ps
}
