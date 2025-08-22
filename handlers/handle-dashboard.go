package handlers

import (
	"etop/templates/layouts"
	"etop/utils"
	"net/http"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	return utils.Render(w, r, layouts.AuthLayout("/sign-in", "/sign-in-card"))
}
