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
	"strings"

	"gorm.io/gorm"
)

func HandleMyTasks(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		page, perPage, sortBy, sortDir := parsePageParams(r)
		saring, filters, cari := saringMyTasks(r, user.ID)
		var total int64
		saring().Count(&total)
		var tasks []models.Task
		saring().
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
			Query:      cari,
			Filters:    filters,
		}
		return layouts.Layout("My Tasks", user, features.MyTasks(tasks, user, pageInfo)).Render(r.Context(), w)
	case http.MethodPost:
		page, perPage, sortBy, sortDir := parsePageParams(r)
		saring, filters, cari := saringMyTasks(r, user.ID)
		var total int64
		saring().Count(&total)
		var tasks []models.Task
		saring().
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
			Query:      cari,
			Filters:    filters,
		}
		// Permintaan dari kotak pencarian, penyaring, pengurutan, dan penomoran
		// halaman hanya menukar daftarnya. Merender ulang seluruh halaman akan
		// ikut menjalankan modal buat tugas yang di dalamnya menarik seluruh
		// baris tabel users dua kali, dan itu mendominasi waktu permintaan.
		if r.URL.Query().Get("part") == "list" {
			return features.MyTasksList(tasks, pageInfo).Render(r.Context(), w)
		}
		return features.MyTasks(tasks, user, pageInfo).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}

// saringMyTasks menyusun kueri daftar tugas beserta pencarian dan penyaring
// yang sedang aktif. Kueri dikembalikan sebagai fungsi agar setiap pemakaian
// mendapat objek baru; satu objek *gorm.DB yang sama tidak boleh dipakai untuk
// Count lalu Find karena syaratnya akan menumpuk.
func saringMyTasks(r *http.Request, userID string) (func() *gorm.DB, map[string]string, string) {
	cari := kataKunci(r)
	filters := map[string]string{
		"status":   strings.TrimSpace(r.URL.Query().Get("status")),
		"priority": strings.TrimSpace(r.URL.Query().Get("priority")),
		"type":     strings.TrimSpace(r.URL.Query().Get("type")),
	}
	for k, v := range filters {
		if v == "" {
			delete(filters, k)
		}
	}

	saring := func() *gorm.DB {
		q := db.PgSql.Model(&models.Task{}).
			Where("user_id = ? OR created_by = ?", userID, userID)
		if cari != "" {
			pola := "%" + strings.ToLower(cari) + "%"
			q = q.Where("LOWER(title) LIKE ? OR LOWER(COALESCE(description,'')) LIKE ?", pola, pola)
		}
		if v := filters["status"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				q = q.Where("status_id = ?", n)
			}
		}
		if v := filters["priority"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				q = q.Where("priority_id = ?", n)
			}
		}
		if v := filters["type"]; v != "" {
			q = q.Where("type = ?", v)
		}
		return q
	}
	return saring, filters, cari
}

// kataKunci mengambil kata kunci pencarian dari alamat permintaan.
func kataKunci(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("q"))
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
