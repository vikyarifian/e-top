package services

import (
	"etop/db"
	"etop/models"
)

func GetProjectStatuses() []models.ProjectStatus {
	projectStatuses := []models.ProjectStatus{}
	db.PgSql.Find(&projectStatuses)

	return projectStatuses
}

func GetProjectMembers(projectID string) []models.User {
	projectMembers := []models.ProjectMember{}
	users := []models.User{}
	db.PgSql.Where("project_id=?", projectID).Preload("User").Find(&projectMembers)
	for _, member := range projectMembers {
		users = append(users, member.User)
	}

	return users
}
