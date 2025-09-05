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
