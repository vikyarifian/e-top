package handlers

import (
	"etop/templates/pages"
	"etop/utils"
	"net/http"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) error {
	return utils.Render(w, r, pages.Dashboard())
}
