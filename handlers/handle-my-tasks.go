package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"net/http"

	"gorm.io/gorm"
)

func HandleMyTasks(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		var tasks []models.Task
		db.PgSql.Where("user_id = ? OR created_by = ?", user.ID, user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order("created_at DESC").
			Find(&tasks)
		return layouts.Layout("My Tasks", user, features.MyTasks(tasks, user)).Render(r.Context(), w)
	case http.MethodPost:
		var tasks []models.Task
		db.PgSql.Where("user_id = ? OR created_by = ?", user.ID, user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order("created_at DESC").
			Find(&tasks)
		return features.MyTasks(tasks, user).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}
