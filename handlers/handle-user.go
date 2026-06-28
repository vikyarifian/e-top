package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func HandleEditUserForm(w http.ResponseWriter, r *http.Request) error {
	user_id := r.URL.Query().Get("user_id")
	var user models.User
	if err := db.PgSql.Where("id=?", user_id).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return ui.Toast("user-error", "warning", "", "User not found!", "", nil).Render(r.Context(), w)
	}
	return features.EditUserForm(user).Render(r.Context(), w)
}

func HandleCreateUserForm(w http.ResponseWriter, r *http.Request) error {
	return features.CreateUserForm().Render(r.Context(), w)
}

func HandleManageUser(w http.ResponseWriter, r *http.Request) error {
	currentUser, _ := auth.GetAuth(w, r)
	if currentUser.Level != "ADMIN" {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("user-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	}

	switch r.Method {
	case http.MethodPost:
		r.ParseForm()
		var newUser models.User
		newUser.FullName = r.FormValue("full_name")
		newUser.Username = r.FormValue("username")
		newUser.Email = r.FormValue("email")
		newUser.Level = r.FormValue("level")
		newUser.Color = r.FormValue("color")
		newUser.Password = r.FormValue("password")

		if len(strings.Trim(newUser.FullName, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(newUser.Username, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Username is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(newUser.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Email is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(newUser.Password, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Password is required!", "", nil).Render(r.Context(), w)
		}

		if len(newUser.Password) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Password must be at least 6 characters", "", nil).Render(r.Context(), w)
		}

		var count int64
		db.PgSql.Where("username=?", newUser.Username).Count(&count)
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "danger", "", "Username already taken!", "", nil).Render(r.Context(), w)
		}

		db.PgSql.Where("email=?", newUser.Email).Count(&count)
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "danger", "", "Email already taken!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(newUser.Color, " ")) < 1 {
			newUser.Color = "bg-gray-500"
		}

		no := 0
		t := time.Now()
		db.PgSql.Table("users").Select("max(no)").Row().Scan(&no)
		newUser.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		newUser.CreatedAt = &t
		newUser.CreatedBy = currentUser.ID
		newUser.UpdatedAt = &t
		newUser.UpdatedBy = currentUser.ID

		hashedPassword, err := utils.HashPassword(newUser.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "error", "", "Failed to hash password!", "", nil).Render(r.Context(), w)
		}
		newUser.Password = hashedPassword

		if err := db.PgSql.Create(&newUser).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(currentUser.ID, "created_user", "User", newUser.ID, map[string]any{
			"description": "created user " + newUser.FullName,
		})

		return ui.Toast("user-success", "success", "", "User created successfully.", "", nil).Render(r.Context(), w)
	case http.MethodPut:
		r.ParseForm()
		userID := r.FormValue("user_id")
		var user models.User
		if err := db.PgSql.Where("id=?", userID).First(&user).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "User is not found!", "", nil).Render(r.Context(), w)
		}

		newUsername := r.FormValue("username")
		if newUsername != user.Username {
			var count int64
			db.PgSql.Where("id<>? AND username=?", userID, newUsername).Count(&count)
			if count > 0 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("user-error", "danger", "", "Username already taken!", "", nil).Render(r.Context(), w)
			}
		}

		updated := false
		logDetails := map[string]any{
			"description": "updated user",
		}

		if strings.Trim(user.FullName, " ") != strings.Trim(r.FormValue("full_name"), " ") {
			logDetails["old_full_name"] = user.FullName
			logDetails["new_full_name"] = r.FormValue("full_name")
			updated = true
		}
		user.FullName = r.FormValue("full_name")

		if strings.Trim(user.Username, " ") != strings.Trim(newUsername, " ") {
			logDetails["old_username"] = user.Username
			logDetails["new_username"] = newUsername
			updated = true
		}
		user.Username = newUsername

		if strings.Trim(user.Color, " ") != strings.Trim(r.FormValue("color"), " ") {
			logDetails["old_color"] = user.Color
			logDetails["new_color"] = r.FormValue("color")
			updated = true
		}
		user.Color = r.FormValue("color")

		if strings.Trim(user.Level, " ") != strings.Trim(r.FormValue("level"), " ") {
			logDetails["old_level"] = user.Level
			logDetails["new_level"] = r.FormValue("level")
			updated = true
		}
		user.Level = r.FormValue("level")

		if len(strings.Trim(user.FullName, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(user.Username, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("user-error", "warning", "", "Username is required!", "", nil).Render(r.Context(), w)
		}

		if err := db.PgSql.Save(&user).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("user-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		if updated {
			services.AddLog(currentUser.ID, "updated_user", "User", user.ID, logDetails)
		}

		return ui.Toast("user-success", "success", "", "User updated successfully.", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
