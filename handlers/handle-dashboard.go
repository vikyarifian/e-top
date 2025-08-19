package handlers

import (
	"etop/templates/layouts"
	"net/http"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, layouts.Dashboard())
}
