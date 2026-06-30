package handlers

import (
	"net/http"
	"strconv"

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
	page, perPage, sortBy, sortDir := parsePageParams(r)
	yearStr := r.URL.Query().Get("year")

	var years []int
	db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND completed_at IS NOT NULL", user.ID).
		Select("DISTINCT EXTRACT(YEAR FROM created_at)::int as year").
		Order("year DESC").
		Pluck("year", &years)

	var total int64
	q := db.PgSql.Model(&models.Task{}).Where("user_id = ? AND completed_at IS NOT NULL", user.ID)
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			q = q.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	q.Count(&total)

	var tasks []models.Task
	tq := db.PgSql.Where("user_id = ? AND completed_at IS NOT NULL", user.ID)
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			tq = tq.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	tq.
		Preload("Assignee", func(db *gorm.DB) *gorm.DB { return db }).
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

	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Achieved", user, features.Achieved(tasks, user, pageInfo, years, yearStr)).Render(r.Context(), w)
	case http.MethodPost:
		return features.Achieved(tasks, user, pageInfo, years, yearStr).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}


