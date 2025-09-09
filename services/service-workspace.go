package services

import (
	"etop/db"
	"etop/models"
)

func GetWorkspaceMembers(workspceID string) []models.User {
	members := []models.WorkspaceMember{}
	users := []models.User{}
	db.PgSql.Where("workspace_id=?", workspceID).Preload("User").Find(&members)
	for _, member := range members {
		users = append(users, member.User)
	}

	return users
}
