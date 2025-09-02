package handlers

import (
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/components"
	"etop/templates/components/ui"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {

	if _, err := auth.GetAuth(w, r); err == nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:
		return layouts.AuthLayout("Sign In", pages.SignIn()).Render(r.Context(), w)
	case http.MethodPost:
		var user models.User

		r.ParseForm()
		user.Email = strings.ToLower(r.FormValue("email"))
		password := r.FormValue("password")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "", "Email not valid!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(password, " ")) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "", "Password must be at least 6 characters!", "", nil).Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "", "Can't find account!", "", nil).Render(r.Context(), w)
		}

		if bcrypt.CompareHashAndPassword([]byte(strings.Trim(user.Password, " ")), []byte(password)) == nil {

			if !user.VerifiedEmail {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("signin-error", "warning", "", "Email verification required! Please check your mail box", "", nil).Render(r.Context(), w)
			}

			tokenString, err := auth.GenerateToken(user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("signin-error", "error", "", "Bad Credentials!", "", nil).Render(r.Context(), w)
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
				return nil
			}

		} else {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signin-error", "warning", "", "Invalid password!", "", nil).Render(r.Context(), w)
		}
	default:
		return nil
	}

}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.GetAuth(w, r); err == nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:
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
			return ui.Toast("signup-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "", "Email not valid!", "", nil).Render(r.Context(), w)
		}

		var count int64
		if db.PgSql.Where("email=? or username=?", user.Email, user.Username).First(&user).Count(&count); count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "", "Email already in use!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(user.Password, " ")) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "", "Password must be at least 6 characters!", "", nil).Render(r.Context(), w)
		}

		if strings.Trim(user.Password, " ") != strings.Trim(pass2, " ") {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("signup-error", "warning", "", "Password not match!", "", nil).Render(r.Context(), w)
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
			return ui.Toast("signup-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		url := os.Getenv("APP_URL")
		if os.Getenv("APP_ENV") == "dev" {
			url += os.Getenv("APP_PORT")
		}
		utils.SendVerificationEmail(user, url+"/verify-email?id="+user.ID)

		return ui.Toast("login-success", "success", "", "Register success! Please check your email for verification.", "", nil).Render(r.Context(), w)
	default:
		return nil
	}

}

func HandleResendVerification(w http.ResponseWriter, r *http.Request) error {

	if _, err := auth.GetAuth(w, r); err == nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:
		return layouts.AuthLayout("Resend Verification", pages.ResendVerification()).Render(r.Context(), w)
	case http.MethodPost:
		var user models.User

		r.ParseForm()
		user.Username = r.FormValue("email")
		user.Email = r.FormValue("email")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("resend-error", "warning", "", "Email not valid!", "", nil).Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			return ui.Toast("resend-success", "success", "", "If your email is registered, you will receive a verification link.", "", nil).Render(r.Context(), w)
		}

		if user.VerifiedEmail {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("resend-error", "warning", "", "Your account already verified. You can sign in to continue.", "", nil).Render(r.Context(), w)
		}

		url := os.Getenv("APP_URL")
		if os.Getenv("APP_ENV") == "dev" {
			url += os.Getenv("APP_PORT")
		}
		utils.SendVerificationEmail(user, url+"/verify-email?id="+user.ID)

		return ui.Toast("resend-success", "success", "", "If your email is registered, you will receive a verification link.", "", nil).Render(r.Context(), w)
	default:
		return nil
	}
}

func HandleForgotPassword(w http.ResponseWriter, r *http.Request) error {

	if _, err := auth.GetAuth(w, r); err == nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:

		var resetPass models.ResetPassword
		param := "Reset"
		token := r.URL.Query().Get("token")
		if token == "" {
			param = "Forgot"
		} else {
			var count int64
			if db.PgSql.Where("token_hash=? AND used=0 AND expires_at > ? ", token, time.Now()).First(&resetPass).Count(&count); count == 0 {
				param = "Error"
			}
		}

		return layouts.AuthLayout(param+" Password", pages.ForgotPassword(param, token)).Render(r.Context(), w)
	case http.MethodPost:
		t := time.Now()
		var user models.User
		var resetPass models.ResetPassword
		var newResetPass models.ResetPassword

		r.ParseForm()
		user.Username = r.FormValue("email")
		user.Email = r.FormValue("email")

		if !utils.IsEmailValidRegex(user.Email) || len(strings.Trim(user.Email, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("forgot-error", "warning", "", "Email not valid!", "", nil).Render(r.Context(), w)
		}

		err := db.PgSql.Where("username=? or email=?", user.Email, user.Email).First(&user).Error
		if err != nil {
			return components.CardStatus(200, "Reset Password Sent", "If your email is registered, you will receive a reset link.", "circle-check-big",
				ui.Button("success", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})).Render(r.Context(), w)
		}

		if err = db.PgSql.Where("email=? AND expires_at > ?", user.Email, time.Now()).Order("expires_at desc").First(&resetPass).Error; err == nil {
			if time.Now().After(resetPass.ExpiresAt) {
				return components.CardStatus(200, "Reset Password Sent", "If your email is registered, you will receive a reset link.", "circle-check-big",
					ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})).Render(r.Context(), w)

			} else {
				return components.CardStatus(500, "Reset Password Failed", "Sorry, we're cannot send reset password at this time, try again later", "circle-x",
					ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})).Render(r.Context(), w)
			}
		} else {
			rawToken := uuid.New().String()
			hashedToken, _ := bcrypt.GenerateFromPassword([]byte(rawToken), bcrypt.DefaultCost)

			newResetPass.Email = user.Email
			newResetPass.TokenHash = string(hashedToken)
			newResetPass.Used = 0
			expiresAt := t.Add(15 * time.Minute)
			newResetPass.ExpiresAt = expiresAt
			newResetPass.CreatedAt = &t
			newResetPass.UpdatedAt = &t

			if err = db.PgSql.Create(&newResetPass).Error; err != nil {
				return components.CardStatus(500, "Reset Password Failed", "Sorry, we're cannot send reset password at this time, try again later", "circle-x",
					ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})).Render(r.Context(), w)
			}

			url := os.Getenv("APP_URL")
			if os.Getenv("APP_ENV") == "dev" {
				url += os.Getenv("APP_PORT")
			}
			utils.ResetPasswordEmail(user, url+"/forgot-password?token="+newResetPass.TokenHash)

			return components.CardStatus(200, "Reset Password Sent", "If your email is registered, you will receive a reset link.", "circle-check-big",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})).Render(r.Context(), w)
		}

	case http.MethodPut:
		var user models.User
		var resetPass models.ResetPassword

		r.ParseForm()
		token := r.FormValue("token")
		user.Password = r.FormValue("password")
		pass2 := r.FormValue("password2")

		if len(strings.Trim(user.Password, " ")) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("forgot-error", "warning", "", "Password must be at least 6 characters!", "", nil).Render(r.Context(), w)
		}

		if strings.Trim(user.Password, " ") != strings.Trim(pass2, " ") {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("forgot-error", "warning", "", "Password not match!", "", nil).Render(r.Context(), w)
		}

		err := db.PgSql.Where("token_hash=? AND used = 0 AND expires_at > ?", token, time.Now()).First(&resetPass).Error
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("resend-error", "warning", "", "Invaid or expired request!", "", nil).Render(r.Context(), w)
		}

		var count int64
		if db.PgSql.Where("email=? or username=?", resetPass.Email, resetPass.Email).First(&user).Count(&count); count == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("forgot-error", "warning", "", "Can't find account!", "", nil).Render(r.Context(), w)
		}

		if bcrypt.CompareHashAndPassword([]byte(strings.Trim(user.Password, " ")), []byte(pass2)) == nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("forgot-error", "warning", "", "New password must be different from the last one!", "", nil).Render(r.Context(), w)
		}

		t := time.Now()
		passHash, _ := utils.HashPassword(pass2)

		user.Password = string(passHash)
		user.UpdatedAt = &t

		err = db.PgSql.Save(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("forgot-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		resetPass.Used = 1
		if err = db.PgSql.Save(&resetPass).Error; err != nil {
			println(err.Error())
		}

		if !user.VerifiedEmail {

			url := os.Getenv("APP_URL")
			if os.Getenv("APP_ENV") == "dev" {
				url += os.Getenv("APP_PORT")
			}
			utils.SendVerificationEmail(user, url+"/verify-email?id="+user.ID)

			return ui.Toast("forgot-success", "success", "", "Reset password success! Please check your email for verification.", "", nil).Render(r.Context(), w)
		}

		return ui.Toast("forgot-success", "success", "", "Reset password success! You can sign in to continue.", "", nil).Render(r.Context(), w)

	default:
		return nil
	}
}

func HandleVerifyEmail(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.GetAuth(w, r); err == nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:

		var count int64
		var user models.User

		id := r.URL.Query().Get("id")

		if db.PgSql.Where("id=?", id).First(&user).Count(&count); count > 0 {
			if user.VerifiedEmail {
				return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
					components.CardStatus(400, "Email Already Verified", "Your account already verified. You can sign in to continue.", "circle-x",
						ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
			}

			t := time.Now()
			user.VerifiedEmail = true
			user.UpdatedAt = &t

			err := db.PgSql.Save(&user).Error
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
					components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
						ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
			}

			return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
				components.CardStatus(200, "Email Verified Successfully", "Your account has been successfully verified. You can now sign in to continue.", "circle-check-big",
					ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
			components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)

	default:
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
	w.WriteHeader(http.StatusSeeOther)
	return nil
}
