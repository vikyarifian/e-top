package services

import (
	"etop/db"
	"etop/models"
	"time"
)

func AddLog(userID string, action string, resourceType string, resourceID string, details map[string]any) error {
	t := time.Now()
	if err := db.PgSql.Create(&models.Log{UserID: userID, Action: action, ResourceType: resourceType, ResourceID: resourceID, Details: details, CreatedAt: &t, CreatedBy: userID, UpdatedAt: &t, UpdatedBy: userID}).Error; err != nil {
		println(err.Error())
		return err
	}
	return nil
}
