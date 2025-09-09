package handlers

import (
	"net/http"

	"etop/auth"
	"etop/templates/layouts"
	"etop/templates/pages"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	authUser, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Dashboard", authUser, pages.Dashboard(authUser)).Render(r.Context(), w)
	case http.MethodPost:
		return pages.Dashboard(authUser).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
