package handlers

import (
	"etop/db"
	"etop/models"
	"etop/templates/features"
	"net/http"
)

func HandleTaskActivities(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		taskActivities := []models.Log{}
		taskID := r.URL.Query().Get("task_id")
		if err := db.PgSql.Where("resource_type='Task' AND resource_id=?", taskID).Preload("User").Order("created_at DESC").Find(&taskActivities).Error; err != nil {
			return features.TaskActivities([]models.Log{}).Render(r.Context(), w)
		}
		return features.TaskActivities(taskActivities).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
