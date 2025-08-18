package handlers

import (
	"etop/templates/pages"
	"net/http"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, pages.Dashboard())
}
