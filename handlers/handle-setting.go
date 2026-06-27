package handlers

import (
	"encoding/json"
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"net/http"
	"time"
)

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
			var depts []models.Department
			db.PgSql.Find(&depts)
			return layouts.Layout("Department Settings", user, features.DeptSettings(user, depts)).Render(r.Context(), w)
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
			var depts []models.Department
			db.PgSql.Find(&depts)
			return features.DeptSettings(user, depts).Render(r.Context(), w)
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
