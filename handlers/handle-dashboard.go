package handlers

import (
	"etop/auth"
	"etop/templates/layouts"
	"etop/templates/pages"
	"net/http"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Dashboard", user, pages.Dashboard(user)).Render(r.Context(), w)
	case http.MethodPost:
		return pages.Dashboard(user).Render(r.Context(), w)
	default:
		return nil
	}
}
