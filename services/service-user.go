package services

import (
	"etop/db"
	"etop/dto"
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

func GetUser(user dto.UserAuth) models.User {
	currentUser := models.User{}
	db.PgSql.Where("id=?", user.ID).First(&currentUser)
	return currentUser
}

func GetNonManager() []models.User {
	nonManagers := []models.User{}
	db.PgSql.Where("id NOT IN (SELECT dept_head_id FROM departments) AND id NOT IN (SELECT user_id FROM department_members)").Find(&nonManagers)
	return nonManagers
}

func GetUserWithoutDept() []models.UserRole {
	users := []models.User{}
	userRole := []models.UserRole{}

	db.PgSql.Where("id NOT IN (SELECT user_id FROM department_members) AND id NOT IN (SELECT dept_head_id FROM departments)").Find(&users)
	for _, user := range users {
		var m models.UserRole
		m.ID = user.ID
		m.FullName = user.FullName
		m.Color = user.Color
		userRole = append(userRole, m)
	}
	return userRole
}
