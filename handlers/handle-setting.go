package handlers

import (
	"encoding/json"
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

func settingsDepts(r *http.Request) ([]models.Department, models.PageInfo) {
	page, perPage, _, _ := parsePageParams(r)
	cari := kataKunci(r)

	// Kueri disusun ulang setiap kali dipakai; satu objek *gorm.DB yang sama
	// tidak boleh dipakai untuk Count lalu Find karena syaratnya menumpuk.
	saring := func() *gorm.DB {
		q := db.PgSql.Model(&models.Department{})
		if cari != "" {
			pola := "%" + strings.ToLower(cari) + "%"
			q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(description,'')) LIKE ?", pola, pola)
		}
		return q
	}

	var total int64
	saring().Count(&total)

	var depts []models.Department
	saring().Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}).Preload("DeptHead").Order("no").
		Limit(perPage).Offset((page - 1) * perPage).
		Find(&depts)

	return depts, models.PageInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
		Query:      cari,
	}
}

func settingsUsers(r *http.Request) ([]models.User, models.PageInfo) {
	page, perPage, _, _ := parsePageParams(r)
	cari := kataKunci(r)

	saring := func() *gorm.DB {
		q := db.PgSql.Model(&models.User{})
		if cari != "" {
			pola := "%" + strings.ToLower(cari) + "%"
			q = q.Where(
				"LOWER(COALESCE(full_name,'')) LIKE ? OR LOWER(COALESCE(username,'')) LIKE ? OR LOWER(COALESCE(email,'')) LIKE ?",
				pola, pola, pola)
		}
		return q
	}

	var total int64
	saring().Count(&total)

	var users []models.User
	saring().Order("no").
		Limit(perPage).Offset((page - 1) * perPage).
		Find(&users)

	return users, models.PageInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
		Query:      cari,
	}
}

func HandleSettings(w http.ResponseWriter, r *http.Request) error {
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "general"
	}
	user, err := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodGet:
		if (user.Level != "ADMIN" && tab != "general" && tab != "notifications" && tab != "security") || err != nil {
			return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
		}

		switch tab {
		case "general":
			return layouts.Layout("General Settings", user, features.GeneralSettings(user)).Render(r.Context(), w)
		case "security":
			return layouts.Layout("Security Settings", user, features.SecuritySettings(user)).Render(r.Context(), w)
		case "notifications":
			var userSettings []models.UserSetting
			db.PgSql.Where("user_id=?", user.ID).Preload("Setting").Find(&userSettings)

			if len(userSettings) == 0 {
				var settings []models.Setting
				db.PgSql.Find(&settings)
				t := time.Now()
				for _, sett := range settings {
					userSettings = append(userSettings, models.UserSetting{SettingID: sett.ID, UserID: user.ID, Value: false, CreatedAt: &t, CreatedBy: user.ID, UpdatedAt: &t, UpdatedBy: user.ID})
				}
				db.PgSql.Create(&userSettings)
			}
			db.PgSql.Where("user_id=?", user.ID).Preload("Setting").Find(&userSettings)
			return layouts.Layout("Notification Settings", user, features.NotificationSettings(user, userSettings)).Render(r.Context(), w)
		case "departments":
			depts, pageInfo := settingsDepts(r)
			return layouts.Layout("Department Settings", user, features.DeptSettings(user, depts, pageInfo)).Render(r.Context(), w)
		case "users":
			users, pageInfo := settingsUsers(r)
			return layouts.Layout("User Settings", user, features.UserSettings(user, users, pageInfo)).Render(r.Context(), w)
		case "task-config":
			return layouts.Layout("Task Config Settings", user,
				features.TaskConfigSettings(user, services.GetTaskPriorities(), services.GetTaskImpacts(), services.GetColorOptions())).Render(r.Context(), w)
		default:
			return features.Settings(tab, user).Render(r.Context(), w)
		}
	case http.MethodPost:
		if (user.Level != "ADMIN" && tab != "general" && tab != "notifications" && tab != "security") || err != nil {
			return pages.Forbidden().Render(r.Context(), w)
		}

		switch tab {
		case "general":
			return features.GeneralSettings(user).Render(r.Context(), w)
		case "security":
			return features.SecuritySettings(user).Render(r.Context(), w)
		case "notifications":
			var userSettings []models.UserSetting
			db.PgSql.Where("user_id=?", user.ID).Preload("Setting").Order("no").Find(&userSettings)

			if len(userSettings) == 0 {
				var settings []models.Setting
				db.PgSql.Order("no").Find(&settings)
				t := time.Now()
				for _, sett := range settings {
					userSettings = append(userSettings, models.UserSetting{SettingID: sett.ID, UserID: user.ID, Value: false, CreatedAt: &t, CreatedBy: user.ID, UpdatedAt: &t, UpdatedBy: user.ID})
				}
				db.PgSql.Create(&userSettings)
			}
			db.PgSql.Where("user_id=?", user.ID).Preload("Setting").Find(&userSettings)
			return features.NotificationSettings(user, userSettings).Render(r.Context(), w)
		case "departments":
			depts, pageInfo := settingsDepts(r)
			return features.DeptSettings(user, depts, pageInfo).Render(r.Context(), w)
		case "users":
			users, pageInfo := settingsUsers(r)
			return features.UserSettings(user, users, pageInfo).Render(r.Context(), w)
		case "task-config":
			return features.TaskConfigSettings(user, services.GetTaskPriorities(),
				services.GetTaskImpacts(), services.GetColorOptions()).Render(r.Context(), w)
		default:
			return features.Settings(tab, user).Render(r.Context(), w)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

}

func HandleNotificationSetting(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodPut:
		payloads := r.FormValue("user_settings")
		var userSettings []models.UserSetting
		err := json.Unmarshal([]byte(payloads), &userSettings)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("setting-error", "danger", "", "Failed to get data payloads!", "", nil).Render(r.Context(), w)
		}

		for _, sett := range userSettings {
			var uSett models.UserSetting
			db.PgSql.Where("user_id=? AND setting_id=?", user.ID, sett.Setting.ID).First(&uSett)
			uSett.Value = sett.Value
			db.PgSql.Save(&uSett)
		}
		// if err := db.PgSql.Save(&userSettings).Error; err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return ui.Toast("setting-error", "danger", "", "Failed to update setting!", "", nil).Render(r.Context(), w)
		// }

		return ui.Toast("setting-success", "success", "", "Setting updated successfully!", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

// HandleTaskConfig menyimpan acuan prioritas dan dampak tugas. Kedua tabel ini
// menentukan nilai tiap tugas, yaitu hasil kali bobot prioritas dan bobot
// dampak, sehingga perubahan di sini langsung mengubah indikator Task Value
// Score pada halaman penilaian.
func HandleTaskConfig(w http.ResponseWriter, r *http.Request) error {
	user, err := auth.GetJwtClaims(w, r)
	if err != nil || user.Level != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return ui.Toast("task-config-error", "danger", "", "You don't have permission to change this setting!", "", nil).Render(r.Context(), w)
	}

	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	var priorities []models.TaskPriority
	if err := json.Unmarshal([]byte(r.FormValue("priorities")), &priorities); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("task-config-error", "danger", "", "Failed to read priority payload!", "", nil).Render(r.Context(), w)
	}
	var impacts []models.TaskImpact
	if err := json.Unmarshal([]byte(r.FormValue("impacts")), &impacts); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("task-config-error", "danger", "", "Failed to read impact payload!", "", nil).Render(r.Context(), w)
	}

	// Kode dan label tidak boleh kosong, sebab keduanya dipakai sebagai
	// penanda pada formulir tugas dan pada berkas bukti penelitian.
	for _, p := range priorities {
		if strings.TrimSpace(p.Priority) == "" || strings.TrimSpace(p.Label) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-config-error", "warning", "", "Priority code and label are required!", "", nil).Render(r.Context(), w)
		}
		if p.Weight < 0 || p.MaxDueMinutes < 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-config-error", "warning", "", "Weight and max due cannot be negative!", "", nil).Render(r.Context(), w)
		}
	}
	for _, im := range impacts {
		if strings.TrimSpace(im.Impact) == "" || strings.TrimSpace(im.Label) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-config-error", "warning", "", "Impact code and label are required!", "", nil).Render(r.Context(), w)
		}
		if im.Weight < 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-config-error", "warning", "", "Weight cannot be negative!", "", nil).Render(r.Context(), w)
		}
	}

	// Seluruh baris disimpan dalam satu transaksi agar tidak ada keadaan
	// setengah tersimpan yang membuat nilai tugas tak menentu.
	err = db.PgSql.Transaction(func(tx *gorm.DB) error {
		for _, p := range priorities {
			kode := strings.ToUpper(strings.TrimSpace(p.Priority))
			if err := tx.Model(&models.TaskPriority{}).Where("no = ?", p.No).Updates(map[string]any{
				"priority":        kode,
				"label":           strings.TrimSpace(p.Label),
				"color":           p.Color,
				"value":           p.Value,
				"level":           p.Level,
				"weight":          p.Weight,
				"max_due_minutes": p.MaxDueMinutes,
			}).Error; err != nil {
				return err
			}
			// Kolom kode pada tabel tugas mengikuti tabel acuannya.
			if err := tx.Model(&models.Task{}).Where("priority_id = ?", p.No).
				Update("priority", kode).Error; err != nil {
				return err
			}
		}
		for _, im := range impacts {
			kode := strings.ToUpper(strings.TrimSpace(im.Impact))
			if err := tx.Model(&models.TaskImpact{}).Where("no = ?", im.No).Updates(map[string]any{
				"impact": kode,
				"label":  strings.TrimSpace(im.Label),
				"color":  im.Color,
				"value":  im.Value,
				"level":  im.Level,
				"weight": im.Weight,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Task{}).Where("impact_id = ?", im.No).
				Update("impact", kode).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("penyimpanan konfigurasi tugas gagal", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return ui.Toast("task-config-error", "danger", "", "Failed to save task configuration!", "", nil).Render(r.Context(), w)
	}

	// Acuan yang tersimpan berubah, maka singgahannya harus dibuang. Tanpa ini
	// bobot dan batas tenggat yang baru tidak akan terpakai sampai aplikasi
	// dijalankan ulang.
	services.BersihkanSinggahanAcuan()

	return ui.Toast("task-config-success", "success", "", "Task configuration saved successfully!", "", nil).Render(r.Context(), w)
}
