package handlers

import (
	"etop/templates/pages"
	"etop/utils"
	"net/http"
)

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {
	return utils.Render(w, r, pages.SignIn())
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	return utils.Render(w, r, pages.SignUp())
}
