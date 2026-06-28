package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/features"
	"net/http"
	"strings"
	"time"
)

func HandleTaskComments(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		taskID := r.URL.Query().Get("task_id")
		var comments []models.Comment
		db.PgSql.Where("task_id=?", taskID).Preload("Author").Order("created_at ASC").Find(&comments)
		return features.TaskComments(comments).Render(r.Context(), w)
	case http.MethodPost:
		user, _ := auth.GetAuth(w, r)
		r.ParseForm()
		text := r.FormValue("text")
		taskID := r.URL.Query().Get("task_id")
		if strings.TrimSpace(text) == "" {
			var comments []models.Comment
			db.PgSql.Where("task_id=?", taskID).Preload("Author").Order("created_at ASC").Find(&comments)
			return features.TaskComments(comments).Render(r.Context(), w)
		}
		t := time.Now().Local()
		comment := models.Comment{
			Text:      text,
			TaskID:    taskID,
			AuthorID:  user.ID,
			CreatedAt: &t,
			CreatedBy: user.ID,
			UpdatedAt: &t,
			UpdatedBy: user.ID,
		}
		if err := db.PgSql.Create(&comment).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
		services.AddLog(user.ID, "added_comment", "Task", taskID, map[string]any{
			"description": "added a comment",
		})
		var comments []models.Comment
		db.PgSql.Where("task_id=?", taskID).Preload("Author").Order("created_at ASC").Find(&comments)
		return features.TaskComments(comments).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
