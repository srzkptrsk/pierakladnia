package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func GlossaryList(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeProject := GetActiveProjectFromContext(r.Context())
		cat := r.URL.Query().Get("category")

		// Pagination parameters
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage < 1 {
			perPage = 25 // App-wide default
		}

		offset := (page - 1) * perPage

		totalItems, err := db.CountGlossaryTerms(deps.DB, activeProject.ID, cat)
		if err != nil {
			http.Error(w, "DB error counting", http.StatusInternalServerError)
			return
		}

		terms, err := db.GetGlossaryTermsPaginated(deps.DB, activeProject.ID, cat, perPage, offset)
		if err != nil {
			http.Error(w, "Failed to fetch glossary", http.StatusInternalServerError)
			return
		}

		totalPages := (totalItems + perPage - 1) / perPage

		render.HTML(w, r, http.StatusOK, "glossary_list.html", map[string]interface{}{
			"Terms":         terms,
			"Category":      cat,
			"Page":          page,
			"PerPage":       perPage,
			"TotalPages":    totalPages,
			"TotalItems":    totalItems,
			"Me":            GetUserFromContext(r.Context()),
			"ActiveProject": activeProject,
			"UserProjects":  GetUserProjectsFromContext(r.Context()),
		})

	}
}

func GlossaryCreate(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if !user.CanTranslate && user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodGet {
			// Show empty form using detail view passing a nil Term
			render.HTML(w, r, http.StatusOK, "glossary_detail.html", map[string]interface{}{
				"Term":          nil,
				"Me":            user,
				"ActiveProject": GetActiveProjectFromContext(r.Context()),
				"UserProjects":  GetUserProjectsFromContext(r.Context()),
			})

			return
		}

		if r.Method == http.MethodPost {
			r.ParseForm()
			cat := r.FormValue("category")
			src := r.FormValue("source_term")
			tgt := r.FormValue("target_term")
			desc := r.FormValue("description")

			var description *string
			if desc != "" {
				description = &desc
			}

			activeProject := GetActiveProjectFromContext(r.Context())
			if err := db.CreateGlossaryTerm(deps.DB, activeProject.ID, cat, src, tgt, description); err != nil {
				http.Error(w, "Failed to create term", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/glossary", http.StatusFound)
		}
	}
}

func GlossaryDetails(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		termID, err := strconv.Atoi(parts[1])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		activeProject := GetActiveProjectFromContext(r.Context())
		term, err := db.GetGlossaryTermByID(deps.DB, activeProject.ID, termID)
		if err != nil || term == nil {
			http.Error(w, "Term not found", http.StatusNotFound)
			return
		}

		user := GetUserFromContext(r.Context())

		if r.Method == http.MethodGet {
			comments, _ := db.GetCommentsForEntity(deps.DB, "glossary_term", termID)
			render.HTML(w, r, http.StatusOK, "glossary_detail.html", map[string]interface{}{
				"Term":          term,
				"Comments":      comments,
				"Me":            user,
				"ActiveProject": activeProject,
				"UserProjects":  GetUserProjectsFromContext(r.Context()),
			})

			return
		}

		// Handle POST update
		if r.Method == http.MethodPost {
			if !user.CanTranslate && user.Role != "admin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			r.ParseForm()
			cat := r.FormValue("category")
			src := r.FormValue("source_term")
			tgt := r.FormValue("target_term")
			desc := r.FormValue("description")

			var description *string
			if desc != "" {
				description = &desc
			}

			if err := db.UpdateGlossaryTerm(deps.DB, activeProject.ID, termID, cat, src, tgt, description); err != nil {
				http.Error(w, "Failed to update term", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/glossary/%d", termID), http.StatusFound)
		}
	}
}
