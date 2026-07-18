package handlers

import (
	"net/http"
	"strconv"

	"etop/auth"
	"etop/db"
	"etop/dto"
	"etop/models"
	"etop/services"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"

	"gorm.io/gorm"
)

const notCancelledFilter = "status_id NOT IN (SELECT no FROM task_statuses WHERE status = 'CANCELLED')"

// achievedViewUsers returns the users whose evaluation the viewer may see:
// admins see every user that has tasks, department heads see their members.
func achievedViewUsers(user dto.UserAuth) []models.User {
	var viewUsers []models.User
	if user.Level == "ADMIN" {
		db.PgSql.
			Where("id IN (SELECT DISTINCT user_id FROM tasks)").
			Order("full_name").
			Find(&viewUsers)
		return viewUsers
	}

	db.PgSql.
		Where(`id IN (
			SELECT dm.user_id FROM department_members dm
			WHERE dm.department_id IN (SELECT d.id FROM departments d WHERE d.dept_head_id = ?)
		) OR id = ?`, user.ID, user.ID).
		Order("full_name").
		Find(&viewUsers)
	if len(viewUsers) <= 1 {
		return nil
	}
	return viewUsers
}

func HandleAchieved(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	page, perPage, sortBy, sortDir := parsePageParams(r)
	yearStr := r.URL.Query().Get("year")

	viewUsers := achievedViewUsers(user)

	targetID := user.ID
	selectedUser := ""
	if requested := r.FormValue("user_id"); requested != "" && requested != user.ID {
		for _, vu := range viewUsers {
			if vu.ID == requested {
				targetID = requested
				selectedUser = requested
				break
			}
		}
	}

	var years []int
	db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND "+notCancelledFilter, targetID).
		Select("DISTINCT EXTRACT(YEAR FROM COALESCE(completed_at, created_at))::int as year").
		Order("year DESC").
		Pluck("year", &years)

	eval := services.GetAchievedEvaluation(targetID, yearStr)

	var total int64
	q := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND "+notCancelledFilter, targetID)
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			q = q.Where("EXTRACT(YEAR FROM COALESCE(completed_at, created_at)) = ?", y)
		}
	}
	q.Count(&total)

	var tasks []models.Task
	tq := db.PgSql.
		Where("user_id = ? AND "+notCancelledFilter, targetID)
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			tq = tq.Where("EXTRACT(YEAR FROM COALESCE(completed_at, created_at)) = ?", y)
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
		return layouts.Layout("Achieved", user, features.Achieved(tasks, user, pageInfo, years, yearStr, eval, viewUsers, selectedUser)).Render(r.Context(), w)
	case http.MethodPost:
		return features.Achieved(tasks, user, pageInfo, years, yearStr, eval, viewUsers, selectedUser).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}
