package services

import (
	"etop/db"
	"etop/models"
)

func GetWorkspaces() []models.Workspace {
	ws := []models.Workspace{}
	ws = append(ws, models.Workspace{ID: "1", Name: "IT Dept"})
	ws = append(ws, models.Workspace{ID: "2", Name: "Legal"})
	ws = append(ws, models.Workspace{ID: "3", Name: "Finance"})

	return ws
}

func GetWorkspaceMembers(workspceID string) []models.User {
	members := []models.WorkspaceMember{}
	users := []models.User{}
	db.PgSql.Where("workspace_id=?", workspceID).Preload("User").Find(&members)
	for _, member := range members {
		users = append(users, member.User)
	}

	return users
}
