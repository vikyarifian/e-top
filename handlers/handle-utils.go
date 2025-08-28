package handlers

import (
	"encoding/json"
	"errors"
	"etop/auth"
	"etop/templates/layouts"
	"etop/templates/pages"
	"log/slog"
	"net/http"
)

type HTTPHandler func(w http.ResponseWriter, r *http.Request) error

func Make(h HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var status int
			switch {
			case errors.Is(err, ErrUnauthorized):
				status = http.StatusUnauthorized
			case errors.Is(err, ErrForbidden):
				status = http.StatusForbidden
			case errors.Is(err, ErrNotFound):
				status = http.StatusNotFound
			default:
				status = http.StatusInternalServerError
			}

			slog.Error("HTTP handler error",
				"err", err,
				"path", r.URL.Path,
				"status", status,
			)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":  err.Error(),
				"status": status,
			})
		}
	}
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
)

func HandleNotFound(w http.ResponseWriter, r *http.Request) error {

	user, err := auth.GetAuth(w, r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusSeeOther)
	}

	switch r.Method {
	case http.MethodGet:
		return layouts.Layout("Not Found", user, pages.NotFound()).Render(r.Context(), w)
	case http.MethodPost:
		return pages.NotFound().Render(r.Context(), w)
	default:
		return nil
	}
}
