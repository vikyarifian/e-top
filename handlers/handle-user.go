package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/utils"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// func HandleUsers(w http.ResponseWriter, r *http.Request) error {

// }

func HandleProfile(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodPut:
		var userExisting models.User
		if err := db.PgSql.Where("id=?", user.ID).First(&userExisting).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "User is not found!", "", nil).Render(r.Context(), w)
		}
		userExisting.FullName = r.FormValue("full_name")
		var newUserName = r.FormValue("username")
		if newUserName != userExisting.Username {
			var count int64
			db.PgSql.Where("id<>? AND username=?", user.ID, newUserName).Count(&count)
			if count > 0 {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("user-error", "danger", "", "Username already taken!", "", nil).Render(r.Context(), w)
			}
		}
		userExisting.Username = newUserName
		if err := db.PgSql.Save(&userExisting).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "danger", "", "Failed to update user!", "", nil).Render(r.Context(), w)
		}
		services.AddLog(userExisting.ID, "updated_user", "User", userExisting.ID, map[string]any{
			"description":   strings.Trim(user.FullName, " ") + " updated their profile information",
			"old_full_name": user.FullName,
			"new_full_name": userExisting.FullName,
		})
		_, err := auth.GenerateToken(userExisting)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "danger", "", "Failed to refresh user session!", "", nil).Render(r.Context(), w)
		}
		// http.SetCookie(w, &http.Cookie{
		// 	Name:   "session_token",
		// 	Value:  "",
		// 	Path:   "/",
		// 	MaxAge: -1,
		// })
		// time.Sleep(1 * time.Second)
		// http.SetCookie(w, &http.Cookie{
		// 	Name:     "session_token",
		// 	Value:    tokenString,
		// 	Path:     "/",
		// 	HttpOnly: true,
		// 	Secure:   false, // set true kalau pakai https
		// 	Expires:  time.Now().Add(24 * time.Hour),
		// })
		return ui.Toast("user-success", "success", "", "User updated successfully!", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleChangePassword(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodPut:
		var userExisting models.User
		if err := db.PgSql.Where("id=?", user.ID).First(&userExisting).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "User is not found!", "", nil).Render(r.Context(), w)
		}
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if len(newPassword) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Password must be at least 6 characters", "", nil).Render(r.Context(), w)
		}

		if newPassword != confirmPassword {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "New & Confirm password is not match!", "", nil).Render(r.Context(), w)
		}
		if newPassword == currentPassword {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Current & New password must be different!", "", nil).Render(r.Context(), w)
		}

		if bcrypt.CompareHashAndPassword([]byte(strings.Trim(userExisting.Password, " ")), []byte(currentPassword)) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Current password is incorrect!", "", nil).Render(r.Context(), w)
		}
		hashedPassword, err := utils.HashPassword(newPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "danger", "", "Failed to create new password!", "", nil).Render(r.Context(), w)
		}
		userExisting.Password = hashedPassword
		if err := db.PgSql.Save(&userExisting).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "danger", "", "Failed to update password!", "", nil).Render(r.Context(), w)
		}
		services.AddLog(userExisting.ID, "updated_user", "User", userExisting.ID, map[string]any{
			"description": strings.Trim(user.FullName, " ") + " updated their password",
		})
		return ui.Toast("user-success", "success", "", "Password updated successfully!", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
