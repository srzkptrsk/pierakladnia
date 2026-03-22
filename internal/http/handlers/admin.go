package handlers

import (
	"fmt"
	"net/http"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func Me(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		render.HTML(w, r, http.StatusOK, "me.html", map[string]interface{}{
			"User": user,
			"Me":   user,
		})

	}
}

func AdminUsersList(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := db.GetAllUsers(deps.DB)
		if err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		projects, err := db.GetAllProjects(deps.DB)
		if err != nil {
			http.Error(w, "Failed to load projects", http.StatusInternalServerError)
			return
		}

		// Build a map of userID -> list of assigned project IDs
		userProjects := make(map[int][]int)
		for _, u := range users {
			ids, _ := db.GetProjectIDsForUser(deps.DB, u.ID)
			userProjects[u.ID] = ids
		}

		render.HTML(w, r, http.StatusOK, "admin_users.html", map[string]interface{}{
			"Users":           users,
			"Projects":        projects,
			"UserProjectsMap": userProjects,
			"Me":              GetUserFromContext(r.Context()),
			"ActiveProject":   GetActiveProjectFromContext(r.Context()),
			"UserProjects":    GetUserProjectsFromContext(r.Context()),
		})

	}
}

func AdminToggleTranslate(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Ensure admin
		adminUser := GetUserFromContext(r.Context())
		if adminUser == nil || adminUser.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		r.ParseForm()
		userIDStr := r.FormValue("user_id")

		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)

		if userID > 0 {
			if err := db.ToggleCanTranslate(deps.DB, userID); err != nil {
				http.Error(w, "Failed to toggle permission", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/admin/users", http.StatusFound)
	}
}
