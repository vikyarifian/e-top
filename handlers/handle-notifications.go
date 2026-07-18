package handlers

import (
	"net/http"
	"strconv"

	"etop/auth"
	"etop/models"
	"etop/services"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
)

func HandleNotifications(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage := 20

	notifs, total := services.GetUserNotificationsPaged(user.ID, perPage, (page-1)*perPage)
	pageInfo := models.PageInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Notifications", user, features.Notifications(notifs, pageInfo)).Render(r.Context(), w)
	case http.MethodPost:
		return features.Notifications(notifs, pageInfo).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return pages.NotFound().Render(r.Context(), w)
	}
}
