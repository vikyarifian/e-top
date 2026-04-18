package services

import (
	"etop/db"
	"etop/models"
)

func GetWorkspaceMembers(workspceID string) []models.UserRole {
	members := []models.WorkspaceMember{}
	users := []models.UserRole{}
	db.PgSql.Where("workspace_id=?", workspceID).Preload("User").Find(&members)
	for _, member := range members {
		var m models.UserRole
		m.ID = member.User.ID
		m.FullName = member.User.FullName
		m.Role = member.Role
		m.Color = member.User.Color
		users = append(users, m)
	}

	return users
}
