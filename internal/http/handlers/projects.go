package handlers

import (
	"net/http"
	"strconv"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func AdminProjectsList(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := db.GetAllProjects(deps.DB)
		if err != nil {
			http.Error(w, "Failed to load projects", http.StatusInternalServerError)
			return
		}

		users, err := db.GetAllUsers(deps.DB)
		if err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		// Also we might need to know who is in what project, but for MVP it might be easier
		// to just show a list of projects and a form to add a user.
		render.HTML(w, http.StatusOK, "admin_projects.html", map[string]interface{}{
			"Projects":      projects,
			"AllUsers":      users, // For the assignment dropdown
			"ActiveProject": GetActiveProjectFromContext(r.Context()),
			"UserProjects":  GetUserProjectsFromContext(r.Context()),
			"Me":            GetUserFromContext(r.Context()),
		})
	}
}

func AdminProjectsCreate(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		name := r.FormValue("name")
		description := r.FormValue("description")

		if name != "" {
			_, err := db.CreateProject(deps.DB, name, description)
			if err != nil {
				http.Error(w, "Failed to create project", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/admin/projects", http.StatusFound)
	}
}

func AdminProjectsEdit(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		projectID, _ := strconv.Atoi(r.FormValue("project_id"))
		name := r.FormValue("name")
		description := r.FormValue("description")

		if projectID > 0 && name != "" {
			err := db.UpdateProject(deps.DB, projectID, name, description)
			if err != nil {
				http.Error(w, "Failed to update project", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/admin/projects", http.StatusFound)
	}
}

func AdminProjectsAssignUser(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		projectID, _ := strconv.Atoi(r.FormValue("project_id"))
		userID, _ := strconv.Atoi(r.FormValue("user_id"))
		action := r.FormValue("action") // "assign" or "remove"

		if projectID > 0 && userID > 0 {
			if action == "remove" {
				err := db.RemoveUserFromProject(deps.DB, projectID, userID)
				if err != nil {
					http.Error(w, "Failed to remove user from project", http.StatusInternalServerError)
					return
				}
			} else {
				err := db.AssignUserToProject(deps.DB, projectID, userID)
				if err != nil {
					http.Error(w, "Failed to assign user to project", http.StatusInternalServerError)
					return
				}
			}
		}

		http.Redirect(w, r, func() string {
			if rd := r.FormValue("redirect"); rd != "" {
				return rd
			}
			return "/admin/projects"
		}(), http.StatusFound)
	}
}

func SwitchProject(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Just redirect back to the referer, the middleware picks up ?project_id=
		// Or we can explicitly set the cookie here if we don't want to rely on middleware query param logic
		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
}
