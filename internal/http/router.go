package app

import (
	"log"
	"net/http"

	"pierakladnia/internal/app"
	"pierakladnia/internal/http/handlers"
	"pierakladnia/internal/render"
)

func NewRouter(deps *app.App) http.Handler {
	mux := http.NewServeMux()

	// Load templates once on boot (or you can reload per request in dev)
	if err := render.LoadTemplates("web/templates/*.html"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Static files (CSS, etc)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Health check
	mux.HandleFunc("/healthz", handlers.Health(deps))

	// Root route - redirect based on auth state
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		cookie, err := r.Cookie(deps.Config.Auth.CookieName)
		if err == nil && cookie.Value != "" {
			http.Redirect(w, r, "/strings", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	// Public Auth
	mux.HandleFunc("/register", handlers.Register(deps))
	mux.HandleFunc("/login", handlers.Login(deps))
	mux.HandleFunc("/logout", handlers.Logout(deps))
	mux.HandleFunc("/verify", handlers.VerifyEmail(deps))

	// Authenticated routes
	mux.Handle("/me", RequireAuth(deps, http.HandlerFunc(handlers.Me(deps))))
	mux.Handle("/strings", RequireAuth(deps, http.HandlerFunc(handlers.StringsList(deps))))
	mux.Handle("/strings/", RequireAuth(deps, http.HandlerFunc(handlers.StringDetails(deps))))
	mux.Handle("/translations/status", RequireAuth(deps, http.HandlerFunc(handlers.TranslationStatusUpdate(deps))))
	mux.Handle("/statistics", RequireAuth(deps, http.HandlerFunc(handlers.ProjectStatistics(deps))))

	mux.Handle("/comments", RequireAuth(deps, http.HandlerFunc(handlers.CreateComment(deps))))

	// Projects
	mux.Handle("/projects/switch", RequireAuth(deps, http.HandlerFunc(handlers.SwitchProject(deps))))

	// Glossary
	mux.Handle("/glossary", RequireAuth(deps, http.HandlerFunc(handlers.GlossaryList(deps))))
	mux.Handle("/glossary/new", RequireAuth(deps, http.HandlerFunc(handlers.GlossaryCreate(deps))))
	// /glossary/{id} handles both view and edit submissions
	mux.Handle("/glossary/", RequireAuth(deps, http.HandlerFunc(handlers.GlossaryDetails(deps))))

	// Admin
	mux.Handle("/admin/users", RequireAdmin(deps, http.HandlerFunc(handlers.AdminUsersList(deps))))
	mux.Handle("/admin/users/toggle-translate", RequireAdmin(deps, http.HandlerFunc(handlers.AdminToggleTranslate(deps))))
	mux.Handle("/admin/projects", RequireAdmin(deps, http.HandlerFunc(handlers.AdminProjectsList(deps))))
	mux.Handle("/admin/projects/create", RequireAdmin(deps, http.HandlerFunc(handlers.AdminProjectsCreate(deps))))
	mux.Handle("/admin/projects/edit", RequireAdmin(deps, http.HandlerFunc(handlers.AdminProjectsEdit(deps))))
	mux.Handle("/admin/projects/assign", RequireAdmin(deps, http.HandlerFunc(handlers.AdminProjectsAssignUser(deps))))
	mux.Handle("/admin/strings/export", RequireAuth(deps, http.HandlerFunc(handlers.AdminStringsExport(deps))))
	mux.Handle("/admin/strings/export/po", RequireAuth(deps, http.HandlerFunc(handlers.AdminStringsExportPO(deps))))
	mux.Handle("/admin/strings/import", RequireAdmin(deps, http.HandlerFunc(handlers.AdminStringsImport(deps))))
	mux.Handle("/admin/strings/import/po", RequireAdmin(deps, http.HandlerFunc(handlers.AdminStringsImportPO(deps))))

	// Apply global middlewares (logging, compression, etc)
	handler := LoggingMiddleware(mux)
	handler = GzipMiddleware(handler)
	return handler
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
