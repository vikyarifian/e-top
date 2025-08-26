package handlers

import (
	"context"
	"encoding/json"
	"etop/auth"
	"etop/db"
	"etop/dto"
	"etop/models"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OAuthConf *oauth2.Config

func OAthConfig() {
	port := strings.Trim(os.Getenv("APP_PORT"), " ")
	if strings.Trim(os.Getenv("APP_ENV"), " ") != "dev" {
		port = ""
	}
	println(strings.Trim(os.Getenv("APP_URL"), " ") + port + strings.Trim(os.Getenv("GOOGLE_REDIRECT_URI"), " "))
	OAuthConf = &oauth2.Config{
		RedirectURL:  strings.Trim(os.Getenv("APP_URL"), " ") + port + strings.Trim(os.Getenv("GOOGLE_REDIRECT_URI"), " "),
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
	url := OAuthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return nil
}

func HandleCallbackGoogle(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return layouts.AuthLayout("Sign In", pages.VerifyEmail("error")).Render(r.Context(), w)
	}

	tokenGoogle, err := OAuthConf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusUnauthorized)
		return layouts.AuthLayout("Sign In", pages.VerifyEmail("error")).Render(r.Context(), w)
	}

	client := OAuthConf.Client(context.Background(), tokenGoogle)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusUnauthorized)
		return layouts.AuthLayout("Sign In", pages.VerifyEmail("error")).Render(r.Context(), w)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var user models.User
	var userInfo struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
		Locale  string `json:"locale"`
	}

	_ = json.Unmarshal(body, &userInfo)

	var count int64
	if db.PgSql.Where("email=? or username=?", userInfo.Email, userInfo.Email).First(&user).Count(&count); count == 0 {
		t := time.Now()
		idHash := utils.GenerateHash(user.Email)
		parts := strings.Split(userInfo.Email, "@")
		passHash, _ := utils.HashPassword(parts[0])

		user.ID = idHash
		user.Username = userInfo.Email
		user.Email = userInfo.Email
		user.FullName = userInfo.Name
		user.Password = string(passHash)
		user.Level = "USER"
		user.VerifiedEmail = true
		user.CreatedAt = &t
		user.CreatedBy = user.ID
		user.UpdatedAt = &t
		user.UpdatedBy = user.ID

		err := db.PgSql.Create(&user).Error
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return layouts.AuthLayout("Sign In", pages.VerifyEmail("error")).Render(r.Context(), w)
		}
	}

	time.Sleep(1 * time.Second)

	tokenString, err := auth.GenerateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return layouts.AuthLayout("Sign In", pages.VerifyEmail("error")).Render(r.Context(), w)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // set true kalau pakai https
			Expires:  time.Now().Add(24 * time.Hour),
		})

		w.Header().Set("HX-Redirect", "/")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		w.WriteHeader(http.StatusOK)
	}
	userAuth := dto.UserAuth{
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
	}

	return layouts.Layout("Dashboard", userAuth, pages.Dashboard(userAuth)).Render(r.Context(), w)
}
