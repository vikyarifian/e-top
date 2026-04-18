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

func GetProjectMembers(projectID string) []models.UserRole {
	projectMembers := []models.ProjectMember{}
	users := []models.UserRole{}
	db.PgSql.Where("project_id=?", projectID).Preload("User").Find(&projectMembers)
	for _, member := range projectMembers {
		var m models.UserRole
		m.ID = member.User.ID
		m.FullName = member.User.FullName
		m.Role = member.Role
		m.Color = member.User.Color
		users = append(users, m)
	}

	return users
}
