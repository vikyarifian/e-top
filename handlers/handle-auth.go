package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/components/ui"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie != nil {
			if _, err := auth.ValidateJWT(cookie.Value); err == nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
		}

		return layouts.AuthLayout("Sign In", pages.SignIn()).Render(r.Context(), w)
	case http.MethodPost:
		var user models.User

		r.ParseForm()
		user.Email = r.FormValue("email")
		password := r.FormValue("password")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Email not valid!").Render(r.Context(), w)
		}

		if len(strings.Trim(password, " ")) < 8 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Password must be at least 8 characters!").Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Can't find account!").Render(r.Context(), w)
		}

		if bcrypt.CompareHashAndPassword([]byte(strings.Trim(user.Password, " ")), []byte(password)) == nil {
			// userAuth := dto.UserAuth{
			// 	Username: user.Username,
			// 	Email:    user.Email,
			// 	FullName: user.FullName,
			// 	Level:    user.Level,
			// 	IsAuth:   true,
			// }

			tokenString, err := auth.GenerateToken(user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("signin-error", "warning", "Error", "Bad Credentials!").Render(r.Context(), w)
			} else {
				// userAuth.Token = tokenString
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    tokenString,
					Path:     "/",
					HttpOnly: true,
					Secure:   false, // set true kalau pakai https
					Expires:  time.Now().Add(24 * time.Hour),
				})

				time.Sleep(1 * time.Second)
				w.Header().Set("HX-Redirect", "/")
				w.WriteHeader(http.StatusOK)
			}

		} else {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Invalid password!").Render(r.Context(), w)
		}
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	w.WriteHeader(http.StatusSeeOther)
	return nil

}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie != nil {
			if _, err := auth.ValidateJWT(cookie.Value); err == nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
		}

		return layouts.AuthLayout("Sign Up", pages.SignUp()).Render(r.Context(), w)
	case http.MethodPost:
		var user models.User

		r.ParseForm()
		user.Username = r.FormValue("email")
		user.Password = r.FormValue("password")
		pass2 := r.FormValue("password2")
		user.Email = r.FormValue("email")
		user.FullName = r.FormValue("full_name")
		user.Level = "USER"

		if len(strings.Trim(user.FullName, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Name is required!").Render(r.Context(), w)
		}

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Email not valid!").Render(r.Context(), w)
		}

		var count int64
		if db.PgSql.Where("email=? or username=?", user.Email, user.Username).First(&user).Count(&count); count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Email already exist!").Render(r.Context(), w)
		}

		if len(strings.Trim(user.Password, " ")) < 8 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Password must be at least 8 characters!").Render(r.Context(), w)
		}

		if strings.Trim(user.Password, " ") != strings.Trim(pass2, " ") {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Password not match!").Render(r.Context(), w)
		}

		t := time.Now()
		idHash := utils.GenerateHash(user.Email)
		passHash, _ := utils.HashPassword(user.Password)

		user.ID = idHash
		user.Password = string(passHash)
		user.CreatedAt = &t
		user.CreatedBy = user.ID
		user.UpdatedAt = &t
		user.UpdatedBy = user.ID

		err := db.PgSql.Create(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "error", "Error", err.Error()).Render(r.Context(), w)
		}

		return ui.Toast("login-success", "success", "Success", "Register success! Go to login form.").Render(r.Context(), w)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

}

func HandleLogout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	time.Sleep(1 * time.Second)
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	return nil
}
