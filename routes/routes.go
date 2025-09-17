package routes

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"etop/auth"
	"etop/handlers"
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
	mux.HandleFunc("/forgot-password", handlers.Make(handlers.HandleForgotPassword))
	mux.HandleFunc("/verify-email", handlers.Make(handlers.HandleVerifyEmail))
	mux.HandleFunc("/logout", handlers.Make(handlers.HandleLogout))

	mux.HandleFunc("/dashboard", auth.RequireAuth(handlers.Make(handlers.HandleDashboard)))

	mux.HandleFunc("/workspace-switcher", auth.RequireAuth(handlers.Make(handlers.HandleWorkspaceSwitcher)))
	mux.HandleFunc("/create-workspace-form", auth.RequireAuth(handlers.Make(handlers.HandleCreateWorkspaceForm)))
	mux.HandleFunc("/workspace", auth.RequireAuth(handlers.Make(handlers.HandleWorkspace)))
	mux.HandleFunc("/workspaces", auth.RequireAuth(handlers.Make(handlers.HandleWorkspaces)))

	mux.HandleFunc("/project", auth.RequireAuth(handlers.Make(handlers.HandleProject)))
	mux.HandleFunc("/projects", auth.RequireAuth(handlers.Make(handlers.HandleProjects)))
	mux.HandleFunc("/create-project-form", auth.RequireAuth(handlers.Make(handlers.HandleCreateProjectForm)))

	mux.HandleFunc("/task", auth.RequireAuth(handlers.Make(handlers.HandleTask)))
	mux.HandleFunc("/tasks", auth.RequireAuth(handlers.Make(handlers.HandleTasks)))
	mux.HandleFunc("/create-task-form", auth.RequireAuth(handlers.Make(handlers.HandleCreateTaskFrom)))
	mux.HandleFunc("/task-watch", auth.RequireAuth(handlers.Make(handlers.HandleTaskWatcher)))

	// mux.HandleFunc("/", auth.RequireAuth(handlers.Make(handlers.HandleNotFound)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			if _, err := auth.GetAuth(w, r); err != nil {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/sign-in")
				} else {
					http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
				}
				return
			}
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

	slog.Info("Server Started", "ListenAddr", listenAddr)

	if err := http.ListenAndServe(listenAddr, corsHandler(mux)); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
