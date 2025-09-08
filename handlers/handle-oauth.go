package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/components"
	"etop/templates/components/ui"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
)

var OAuthConf *oauth2.Config

func OAthConfig() {
	appPort := strings.Trim(os.Getenv("APP_PORT"), " ")
	if strings.Trim(os.Getenv("APP_ENV"), " ") != "dev" {
		appPort = ""
	}

	OAuthConf = &oauth2.Config{
		RedirectURL:  strings.Trim(os.Getenv("APP_URL"), " ") + appPort + strings.Trim(os.Getenv("GOOGLE_REDIRECT_URI"), " "),
		ClientID:     strings.Trim(os.Getenv("GOOGLE_CLIENT_ID"), " "),
		ClientSecret: strings.Trim(os.Getenv("GOOGLE_CLIENT_SECRET"), " "),
		Scopes: []string{
			"openid",
			"email",
			"profile"},
		Endpoint: google.Endpoint,
	}
}

func HandleLoginGoogle(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.GetAuth(w, r); err == nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/dashboard")
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
	authUrl := OAuthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
	return nil
}

func HandleCallbackGoogle(w http.ResponseWriter, r *http.Request) error {

	if _, err := auth.GetAuth(w, r); err == nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/dashboard")
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
			components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
	}

	tokenGoogle, err := OAuthConf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusUnauthorized)
		return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
			components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
	}

	client := OAuthConf.Client(context.Background(), tokenGoogle)
	response, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusUnauthorized)
		return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
			components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var user models.User
	var userGoogle struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
		Locale  string `json:"locale"`
	}

	_ = json.Unmarshal(body, &userGoogle)

	var count int64
	t := time.Now()
	if db.PgSql.Where("email=? or username=?", userGoogle.Email, userGoogle.Email).First(&user).Count(&count); count == 0 {
		idHash := utils.GenerateHash(userGoogle.Email)
		splitEmail := strings.Split(userGoogle.Email, "@")
		passwordHash, _ := utils.HashPassword(splitEmail[0])

		user.ID = idHash
		user.Username = userGoogle.Email
		user.Email = userGoogle.Email
		user.FullName = userGoogle.Name
		user.Password = string(passwordHash)
		user.Color = utils.TailwindForUsername(user.Username)
		user.Level = "USER"
		user.VerifiedEmail = true
		user.CreatedAt = &t
		user.CreatedBy = user.ID
		user.UpdatedAt = &t
		user.UpdatedBy = user.ID

		err := db.PgSql.Create(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return layouts.AuthLayout("Verify Email", pages.VerifyEmail(
				components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
					ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)
		}
	} else {
		t := time.Now()
		user.VerifiedEmail = true
		user.UpdatedAt = &t
		user.UpdatedBy = user.ID

		db.PgSql.Save(&user)
	}

	time.Sleep(1 * time.Second)

	tokenString, err := auth.GenerateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return layouts.AuthLayout("Error", pages.VerifyEmail(
			components.CardStatus(500, "Something went wrong", "Sorry, we're cannot verify your email at this time, try again later.", "circle-x",
				ui.Button("back", "outline", "", "", "Back to Sign In", "", templ.Attributes{"onclick": "window.location.href='/sign-in'"})))).Render(r.Context(), w)

	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // set true if use https
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusFound)

		return nil

	}

}
