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

// pilihTargetPenilaian menentukan karyawan yang datanya ditampilkan.
//
// Bila pengguna memilih seseorang lewat penyaring, pilihan itu dipakai selama
// namanya memang ada pada daftar yang boleh ia lihat. Bila belum memilih,
// yang dipakai adalah nama teratas pada daftar, yang sudah terurut menurut
// abjad pada achievedViewUsers. Sebelumnya yang dipakai selalu pengguna yang
// sedang masuk, sehingga atasan yang menilai banyak orang selalu melihat
// dirinya sendiri lebih dulu.
//
// Pengguna yang daftarnya kosong, yaitu karyawan biasa tanpa anggota, tetap
// melihat datanya sendiri.
func pilihTargetPenilaian(user dto.UserAuth, viewUsers []models.User, requested string) (string, string) {
	if requested != "" {
		for _, vu := range viewUsers {
			if vu.ID == requested {
				return vu.ID, vu.ID
			}
		}
	}
	if len(viewUsers) > 0 {
		return viewUsers[0].ID, viewUsers[0].ID
	}
	return user.ID, ""
}

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

	targetID, selectedUser := pilihTargetPenilaian(user, viewUsers, r.FormValue("user_id"))

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
