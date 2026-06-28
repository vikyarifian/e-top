package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

func HandleMyTasks(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		page, perPage, sortBy, sortDir := parsePageParams(r)
		var total int64
		db.PgSql.Model(&models.Task{}).Where("user_id = ? OR created_by = ?", user.ID, user.ID).Count(&total)
		var tasks []models.Task
		db.PgSql.Where("user_id = ? OR created_by = ?", user.ID, user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order(sortBy + " " + sortDir).
			Limit(perPage).Offset((page - 1) * perPage).
			Find(&tasks)
		pageInfo := models.PageInfo{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
			SortBy:     sortBy,
			SortDir:    sortDir,
		}
		return layouts.Layout("My Tasks", user, features.MyTasks(tasks, user, pageInfo)).Render(r.Context(), w)
	case http.MethodPost:
		page, perPage, sortBy, sortDir := parsePageParams(r)
		var total int64
		db.PgSql.Model(&models.Task{}).Where("user_id = ? OR created_by = ?", user.ID, user.ID).Count(&total)
		var tasks []models.Task
		db.PgSql.Where("user_id = ? OR created_by = ?", user.ID, user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db
			}).
			Preload("Status").
			Preload("Priority").
			Order(sortBy + " " + sortDir).
			Limit(perPage).Offset((page - 1) * perPage).
			Find(&tasks)
		pageInfo := models.PageInfo{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
			SortBy:     sortBy,
			SortDir:    sortDir,
		}
		return features.MyTasks(tasks, user, pageInfo).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}

func parsePageParams(r *http.Request) (page int, perPage int, sortBy string, sortDir string) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage = 20
	sortBy = r.URL.Query().Get("sort_by")
	sortColumns := map[string]string{
		"title": "title", "type": "type", "status": "status_id",
		"priority": "priority_id", "assignee": "user_id",
		"created_at": "created_at", "due_date": "due_date",
	}
	if col, ok := sortColumns[sortBy]; ok {
		sortBy = col
	} else {
		sortBy = "created_at"
	}
	sortDir = r.URL.Query().Get("sort_dir")
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}
	return
}
