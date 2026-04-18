package services

import (
	"etop/db"
	"etop/models"
)

func GetAllUser() []models.User {
	users := []models.User{}
	db.PgSql.Find(&users)

	return users
}

func GetAllUserRole() []models.UserRole {
	users := []models.User{}
	userRole := []models.UserRole{}

	db.PgSql.Find(&users)
	for _, user := range users {
		var m models.UserRole
		m.ID = user.ID
		m.FullName = user.FullName
		m.Color = user.Color
		userRole = append(userRole, m)
	}
	return userRole
}
