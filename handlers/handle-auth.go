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
	"os"
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
		user.Email = strings.ToLower(r.FormValue("email"))
		password := r.FormValue("password")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Email not valid!").Render(r.Context(), w)
		}

		if len(strings.Trim(password, " ")) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Password must be at least 6 characters!").Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "Error", "Can't find account!").Render(r.Context(), w)
		}

		if bcrypt.CompareHashAndPassword([]byte(strings.Trim(user.Password, " ")), []byte(password)) == nil {

			if !user.VerifiedEmail {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("signin-error", "warning", "Error", "Email verification required! Please check your mail box").Render(r.Context(), w)
			}

			tokenString, err := auth.GenerateToken(user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("signin-error", "error", "Error", "Bad Credentials!").Render(r.Context(), w)
			} else {
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
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
		w.Write([]byte(`<script>window.location.href = "/";</script>`))
		return nil
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte(`<script>window.location.href = "/";</script>`))
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
		user.Username = strings.ToLower(r.FormValue("email"))
		user.Password = r.FormValue("password")
		pass2 := r.FormValue("password2")
		user.Email = strings.ToLower(r.FormValue("email"))
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
			return ui.Toast("signup-error", "warning", "Error", "Email already in use!").Render(r.Context(), w)
		}

		if len(strings.Trim(user.Password, " ")) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "Error", "Password must be at least 6 characters!").Render(r.Context(), w)
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
		user.VerifiedEmail = false
		user.CreatedAt = &t
		user.CreatedBy = user.ID
		user.UpdatedAt = &t
		user.UpdatedBy = user.ID

		err := db.PgSql.Create(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "error", "Error", err.Error()).Render(r.Context(), w)
		}

		url := os.Getenv("APP_URL")
		if os.Getenv("APP_ENV") == "dev" {
			url += os.Getenv("APP_PORT")
		}
		utils.SendVerificationEmail(user, url+"/verify-email?id="+user.ID)

		return ui.Toast("login-success", "success", "Success", "Register success! Please check your email for verification.").Render(r.Context(), w)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

}

func HandleResendVerification(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie != nil {
			if _, err := auth.ValidateJWT(cookie.Value); err == nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
		}
		return layouts.AuthLayout("Resend Verification", pages.ResendVerification()).Render(r.Context(), w)
	case http.MethodPost:
		var user models.User

		r.ParseForm()
		user.Username = r.FormValue("email")
		user.Email = r.FormValue("email")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("resend-error", "warning", "Error", "Email not valid!").Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("resend-error", "warning", "Error", "Can't find account!").Render(r.Context(), w)
		}

		if user.VerifiedEmail {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("resend-error", "warning", "Error", "Your account already verified. You can sign in to continue.").Render(r.Context(), w)
		}

		url := os.Getenv("APP_URL")
		if os.Getenv("APP_ENV") == "dev" {
			url += os.Getenv("APP_PORT")
		}
		utils.SendVerificationEmail(user, url+"/verify-email?id="+user.ID)

		return ui.Toast("resend-success", "success", "Success", "Email resend! Please check your email for verification.").Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusSeeOther)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
}

func HandleVerifyEmail(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie != nil {
			if _, err := auth.ValidateJWT(cookie.Value); err == nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil
			}
		}

		var count int64
		var user models.User

		id := r.URL.Query().Get("id")

		if db.PgSql.Where("id=?", id).First(&user).Count(&count); count > 0 {
			if user.VerifiedEmail {
				return layouts.AuthLayout("Verify Email", pages.VerifyEmail("verified")).Render(r.Context(), w)
			}

			t := time.Now()
			user.VerifiedEmail = true
			user.UpdatedAt = &t

			err := db.PgSql.Save(&user).Error
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return layouts.AuthLayout("Verify Email", pages.VerifyEmail("error")).Render(r.Context(), w)
			}

			return layouts.AuthLayout("Verify Email", pages.VerifyEmail("success")).Render(r.Context(), w)
		}

		w.WriteHeader(http.StatusSeeOther)
		return layouts.AuthLayout("Verify Email", pages.VerifyEmail("not-found")).Render(r.Context(), w)

	default:
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
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
