package handlers

import (
	"net/http"

	"etop/auth"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
)

func HandleNotifications(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Notifications", user, features.Notifications(user.ID)).Render(r.Context(), w)
	case http.MethodPost:
		return features.Notifications(user.ID).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}
