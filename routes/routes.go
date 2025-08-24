package routes

import (
	"etop/auth"
	"etop/handlers"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// Static file handler
func HandleStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/public/")

	// Set appropriate content type
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	http.ServeFile(w, r, "./public/"+path)
}

func SetRoutes() {

	r := http.NewServeMux()

	r.HandleFunc("/public/", HandleStatic)
	// r.Handle("/favicon.ico", http.FileServer(http.Dir("public")))
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	})

	r.HandleFunc("/sign-in", handlers.Make(handlers.HandleSignIn))
	r.HandleFunc("/sign-up", handlers.Make(handlers.HandleSignUp))
	r.HandleFunc("/logout", handlers.Make(handlers.HandleLogout))

	r.HandleFunc("/", auth.RequireAuth(handlers.Make(handlers.HandleDashboard)))
	r.HandleFunc("/workspace-switcher", auth.RequireAuth(handlers.Make(handlers.WorkspaceSwitcher)))

	// Add CORS headers
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	listenAddr := os.Getenv("APP_PORT")

	slog.Info("Server started", "listenAddr", listenAddr)

	if err := http.ListenAndServe(listenAddr, corsHandler(r)); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
