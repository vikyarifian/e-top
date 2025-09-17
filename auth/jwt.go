package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"etop/dto"
	"etop/models"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

// GenerateToken for JWT user
func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":        user.ID,
		"username":  user.Username,
		"full_name": user.FullName,
		"email":     user.Email,
		"level":     user.Level,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"sub":       user.Username + "_" + user.Level,
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Get User Auth from Token
func GetAuth(w http.ResponseWriter, r *http.Request) (dto.UserAuth, error) {
	_, err := r.Cookie("session_token")
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/sign-in")
		} else {
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		}
		return dto.UserAuth{}, err
	}
	authUser := dto.UserAuth{
		ID:       r.Header.Get("X-User-ID"),
		Username: r.Header.Get("X-Username"),
		FullName: r.Header.Get("X-Full-Name"),
		Email:    r.Header.Get("X-Email"),
		Level:    r.Header.Get("X-Level"),
		IsAuth:   true,
		Token:    r.Header.Get("session_token"),
	}

	return authUser, nil
}

func GetJwtClaims(w http.ResponseWriter, r *http.Request) (dto.UserAuth, error) {

	var user dto.UserAuth

	cookie, err := r.Cookie("session_token")
	if err != nil {
		return user, err
	}

	claims, err := ValidateJWT(cookie.Value)
	if err != nil {
		return user, err
	}

	user = dto.UserAuth{
		ID:       fmt.Sprintf("%s", claims["id"]),
		Username: fmt.Sprintf("%s", claims["username"]),
		Email:    fmt.Sprintf("%s", claims["email"]),
		FullName: cases.Title(language.English, cases.Compact).String(fmt.Sprintf("%s", claims["full_name"])),
		Level:    fmt.Sprintf("%s", claims["level"]),
		Token:    cookie.Value,
		IsAuth:   true,
	}

	return user, nil
}

// ParseToken for validate JWT
func ValidateJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, err
}

// RequireAuth middleware for route protection
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/sign-in")
			} else {
				http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			}
			return
		}

		claims, err := ValidateJWT(cookie.Value)
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/sign-in")
			} else {
				http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			}
			return
		}

		// save userAuth to request header
		user := dto.UserAuth{
			ID:       fmt.Sprintf("%s", claims["id"]),
			Username: fmt.Sprintf("%s", claims["username"]),
			Email:    fmt.Sprintf("%s", claims["email"]),
			FullName: cases.Title(language.English, cases.Compact).String(fmt.Sprintf("%s", claims["full_name"])),
			Level:    fmt.Sprintf("%s", claims["level"]),
		}

		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-Username", user.Username)
		r.Header.Set("X-Full-Name", user.FullName)
		r.Header.Set("X-Email", user.Email)
		r.Header.Set("X-Level", user.Level)

		next.ServeHTTP(w, r)
	}
}
