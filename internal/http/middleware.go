package app

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/http/handlers"
	"pierakladnia/internal/render"
)

// RequireAuth ensures the user is logged in
func RequireAuth(deps *app.App, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(deps.Config.Auth.CookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := deps.Sessions.GetUserFromSession(cookie.Value)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			// Invalid or expired session
			http.SetCookie(w, &http.Cookie{
				Name:   deps.Config.Auth.CookieName,
				Value:  "",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Resolve User Projects
		projects, err := db.GetProjectsForUser(deps.DB, user.ID, user.Role)
		if err != nil {
			http.Error(w, "Failed to load projects", http.StatusInternalServerError)
			return
		}

		if len(projects) == 0 {
			render.HTML(w, r, http.StatusOK, "no_projects.html", map[string]interface{}{
				"Me": user,
			})
			return
		}

		// Resolve Active Project
		var activeProject *db.Project

		// 1. Check query param
		if pidStr := r.URL.Query().Get("project_id"); pidStr != "" {
			if pid, err := strconv.Atoi(pidStr); err == nil {
				for _, p := range projects {
					if p.ID == pid {
						activeProject = p
						break
					}
				}
				if activeProject != nil {
					// Set cookie so they remember
					http.SetCookie(w, &http.Cookie{
						Name:    "active_project_id",
						Value:   strconv.Itoa(activeProject.ID),
						Path:    "/",
						Expires: time.Now().Add(30 * 24 * time.Hour),
					})
				}
			}
		}

		// 2. Check cookie if still nil
		if activeProject == nil {
			if pc, err := r.Cookie("active_project_id"); err == nil {
				if pid, err := strconv.Atoi(pc.Value); err == nil {
					for _, p := range projects {
						if p.ID == pid {
							activeProject = p
							break
						}
					}
				}
			}
		}

		// 3. Fallback to first project
		if activeProject == nil {
			activeProject = projects[0]
			http.SetCookie(w, &http.Cookie{
				Name:    "active_project_id",
				Value:   strconv.Itoa(activeProject.ID),
				Path:    "/",
				Expires: time.Now().Add(30 * 24 * time.Hour),
			})
		}

		// Store in context
		ctx := context.WithValue(r.Context(), handlers.UserContextKey, user)
		ctx = context.WithValue(ctx, handlers.UserProjectsContextKey, projects)
		ctx = context.WithValue(ctx, handlers.ActiveProjectContextKey, activeProject)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin ensures the user is logged in and is an admin
func RequireAdmin(deps *app.App, next http.Handler) http.Handler {
	return RequireAuth(deps, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := handlers.GetUserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}
