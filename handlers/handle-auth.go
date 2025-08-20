package handlers

import (
	"etop/templates/layouts"
	"etop/utils"
	"net/http"
)

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {
	return utils.Render(w, r, layouts.SignIn())
}
