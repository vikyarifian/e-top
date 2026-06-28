package handlers

import (
	"net/http"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"

	"gorm.io/gorm"
)

func HandleAchieved(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		var tasks []models.Task
		db.PgSql.Where("user_id = ? AND completed_at IS NOT NULL", user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order("completed_at DESC").
			Find(&tasks)
		return layouts.Layout("Achieved", user, features.Achieved(tasks, user)).Render(r.Context(), w)
	case http.MethodPost:
		var tasks []models.Task
		db.PgSql.Where("user_id = ? AND completed_at IS NOT NULL", user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order("completed_at DESC").
			Find(&tasks)
		return features.Achieved(tasks, user).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}
