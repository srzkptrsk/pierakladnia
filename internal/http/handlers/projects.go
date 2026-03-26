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

		render.HTML(w, r, http.StatusOK, "admin_projects.html", map[string]interface{}{
			"Projects":      projects,
			"AllUsers":      users,
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
		exportFilename := r.FormValue("export_filename")
		poProjectIdVersion := r.FormValue("po_project_id_version")
		poReportMsgidBugsTo := r.FormValue("po_report_msgid_bugs_to")
		poLanguageTeam := r.FormValue("po_language_team")
		poLanguage := r.FormValue("po_language")
		poLastTranslator := r.FormValue("po_last_translator")

		if projectID > 0 && name != "" {
			err := db.UpdateProject(deps.DB, projectID, name, description, exportFilename, poProjectIdVersion, poReportMsgidBugsTo, poLanguageTeam, poLanguage, poLastTranslator)
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
