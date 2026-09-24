package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"etop/db"
	"etop/dto"
	"etop/models"

	"gorm.io/gorm"
)

// deptTasks menyusun daftar tugas seluruh anggota sebuah departemen beserta
// pencarian dan penyaring yang sedang berlaku.
//
// Cakupannya sengaja ditentukan dari keanggotaan departemen, bukan dari peran
// pengguna. Halaman departemen hanya dirender bagi anggotanya, sehingga daftar
// ini tidak pernah memperlihatkan tugas di luar departemen yang sedang dibuka.
func deptTasks(r *http.Request, dept models.Department) dto.DeptTasks {
	page, perPage, sortBy, sortDir := parsePageParams(r)
	cari := kataKunci(r)

	anggota := make([]string, 0, len(dept.Members))
	for _, m := range dept.Members {
		if m.UserID != "" {
			anggota = append(anggota, m.UserID)
		}
	}

	filters := map[string]string{
		"status":   strings.TrimSpace(r.URL.Query().Get("status")),
		"priority": strings.TrimSpace(r.URL.Query().Get("priority")),
		"impact":   strings.TrimSpace(r.URL.Query().Get("impact")),
		"member":   strings.TrimSpace(r.URL.Query().Get("member")),
	}
	for k, v := range filters {
		if v == "" {
			delete(filters, k)
		}
	}

	hasil := dto.DeptTasks{
		DeptID: dept.ID,
		Page: models.PageInfo{
			Page:    page,
			PerPage: perPage,
			SortBy:  sortBy,
			SortDir: sortDir,
			Query:   cari,
			Filters: filters,
		},
	}
	if len(anggota) == 0 {
		return hasil
	}

	// Daftar anggota untuk pilihan penyaring, terurut menurut nama.
	db.PgSql.Model(&models.User{}).
		Where("id IN (?)", anggota).
		Order("full_name").
		Find(&hasil.Members)

	// Penyaring anggota hanya dihormati bila orangnya memang di departemen ini.
	orang := anggota
	if v := filters["member"]; v != "" {
		cocok := false
		for _, id := range anggota {
			if id == v {
				cocok = true
				break
			}
		}
		if cocok {
			orang = []string{v}
		}
	}

	// Kueri disusun ulang setiap kali dipakai; satu objek *gorm.DB yang sama
	// tidak boleh dipakai untuk Count lalu Find karena syaratnya menumpuk.
	saring := func() *gorm.DB {
		q := db.PgSql.Model(&models.Task{}).Where("user_id IN (?)", orang)
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
		if v := filters["impact"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				q = q.Where("impact_id = ?", n)
			}
		}
		return q
	}

	var total int64
	saring().Count(&total)
	saring().
		Preload("Assignee", func(db *gorm.DB) *gorm.DB { return db }).
		Preload("Status").
		Preload("Priority").
		Order(sortBy + " " + sortDir).
		Limit(perPage).Offset((page - 1) * perPage).
		Find(&hasil.Tasks)

	hasil.Page.Total = total
	hasil.Page.TotalPages = int((total + int64(perPage) - 1) / int64(perPage))
	return hasil
}
