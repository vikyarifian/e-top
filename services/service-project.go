package services

import (
	"etop/db"
	"etop/models"
)

func GetProjectStatuses() []models.ProjectStatus {
	ps := []models.ProjectStatus{}
	db.PgSql.Find(&ps)

	return ps
}

func GetProjectMembers(projectID string) []models.User {
	members := []models.ProjectMember{}
	users := []models.User{}
	db.PgSql.Where("project_id=?", projectID).Preload("User").Find(&members)
	for _, member := range members {
		users = append(users, member.User)
	}

	return users
}
