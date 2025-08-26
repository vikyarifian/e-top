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

// helper buat cek status
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

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

	mux := http.NewServeMux()

	mux.HandleFunc("/public/", HandleStatic)
	// mux.Handle("/favicon.ico", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	})

	mux.HandleFunc("/sign-in", handlers.Make(handlers.HandleSignIn))
	mux.HandleFunc("/sign-up", handlers.Make(handlers.HandleSignUp))
	mux.HandleFunc("/auth/google", handlers.Make(handlers.HandleLoginGoogle))
	mux.HandleFunc("/auth/google/callback", handlers.Make(handlers.HandleCallbackGoogle))
	mux.HandleFunc("/resend-verification", handlers.Make(handlers.HandleResendVerification))
	mux.HandleFunc("/send-reset-password", handlers.Make(handlers.HandleResendVerification))
	mux.HandleFunc("/verify-email", handlers.Make(handlers.HandleVerifyEmail))
	mux.HandleFunc("/logout", handlers.Make(handlers.HandleLogout))

	mux.HandleFunc("/dashboard", auth.RequireAuth(handlers.Make(handlers.HandleDashboard)))
	mux.HandleFunc("/workspace-switcher", auth.RequireAuth(handlers.Make(handlers.WorkspaceSwitcher)))

	// mux.HandleFunc("/", auth.RequireAuth(handlers.Make(handlers.HandleNotFound)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			auth.RequireAuth(handlers.Make(handlers.HandleDashboard))(w, r)
			return
		}
		auth.RequireAuth(handlers.Make(handlers.HandleNotFound))(w, r)
	})

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

	if err := http.ListenAndServe(listenAddr, corsHandler(mux)); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
