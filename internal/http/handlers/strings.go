package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func StringsList(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeProject := GetActiveProjectFromContext(r.Context())

		// Pagination parameters
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage < 1 {
			perPage = 25 // Default
		}

		offset := (page - 1) * perPage
		query := strings.ToLower(r.URL.Query().Get("q"))
		status := strings.ToLower(r.URL.Query().Get("status"))
		sortParam := r.URL.Query().Get("sort")

		totalItems, err := db.CountStrings(deps.DB, activeProject.ID, query, status)
		if err != nil {
			http.Error(w, "DB error counting", http.StatusInternalServerError)
			return
		}

		strs, err := db.GetStringsPaginated(deps.DB, activeProject.ID, query, status, sortParam, perPage, offset)
		if err != nil {
			http.Error(w, "DB error fetching", http.StatusInternalServerError)
			return
		}

		totalPages := (totalItems + perPage - 1) / perPage

		render.HTML(w, http.StatusOK, "strings_list.html", map[string]interface{}{
			"Strings":       strs,
			"Query":         query,
			"Status":        status,
			"Sort":          sortParam,
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

func StringDetails(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// URL: /strings/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		stringID, err := strconv.Atoi(parts[1])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		activeProject := GetActiveProjectFromContext(r.Context())
		str, err := db.GetStringByID(deps.DB, activeProject.ID, stringID)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		if str == nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		user := GetUserFromContext(r.Context())

		// Handle POST to update translation or context
		if r.Method == http.MethodPost {
			if !user.CanTranslate && user.Role != "admin" {
				http.Error(w, "Forbidden - Translation disabled", http.StatusForbidden)
				return
			}
			r.ParseForm()

			if len(parts) >= 3 && parts[2] == "context" {
				contextText := r.FormValue("context")
				err = db.UpdateStringContext(deps.DB, activeProject.ID, stringID, contextText)
				if err != nil {
					http.Error(w, "Failed to update context", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, fmt.Sprintf("/strings/%d", stringID), http.StatusFound)
				return
			}

			// Otherwise, it's a translation update
			locale := r.FormValue("locale")
			if locale == "" {
				locale = "target" // Default MVP locale
			}
			newText := r.FormValue("translation")

			err = db.UpsertTranslation(deps.DB, stringID, locale, newText, user.ID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save translation: %v", err), http.StatusInternalServerError)
				return
			}

			// Stay on the same page
			http.Redirect(w, r, fmt.Sprintf("/strings/%d", stringID), http.StatusFound)
			return
		}

		// GET Details page
		translations, _ := db.GetTranslationsForString(deps.DB, stringID)
		comments, _ := db.GetCommentsForEntity(deps.DB, "translation", stringID)

		// Preload revisions for translations
		var revisions []db.TranslationRevision
		if len(translations) > 0 {
			// for MVP we only show revisions for the first locale
			revisions, _ = db.GetRevisionsForTranslation(deps.DB, translations[0].ID)
		}

		render.HTML(w, http.StatusOK, "string_detail.html", map[string]interface{}{
			"String":        str,
			"Translations":  translations,
			"Revisions":     revisions,
			"Comments":      comments,
			"Me":            user,
			"ActiveProject": activeProject,
			"UserProjects":  GetUserProjectsFromContext(r.Context()),
		})
	}
}

func TranslationStatusUpdate(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user := GetUserFromContext(r.Context())
		if !user.CanTranslate && user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		r.ParseForm()
		stringID, _ := strconv.Atoi(r.FormValue("string_id"))
		locale := r.FormValue("locale")
		status := r.FormValue("status")

		if err := db.UpdateTranslationStatus(deps.DB, stringID, locale, status); err != nil {
			http.Error(w, "Failed to update status", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/strings/%d", stringID), http.StatusFound)
	}
}

func AdminStringsImport(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user := GetUserFromContext(r.Context())
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		activeProject := GetActiveProjectFromContext(r.Context())
		if activeProject == nil {
			http.Error(w, "No active project", http.StatusBadRequest)
			return
		}

		r.ParseForm()
		jsonContent := r.FormValue("json_content")
		if jsonContent == "" {
			http.Error(w, "No JSON content provided", http.StatusBadRequest)
			return
		}

		var data map[string]string
		if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
			http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
			return
		}

		for sourceText, targetText := range data {
			// Create the string in the active project
			stringID, err := db.CreateString(deps.DB, activeProject.ID, "", sourceText, "")
			if err != nil {
				continue // Skip if error (e.g. creating string fails)
			}

			// Add the target translation
			if targetText != "" {
				err = db.UpsertTranslation(deps.DB, stringID, "target", targetText, user.ID)
				if err != nil {
					continue
				}
			}
		}

		http.Redirect(w, r, "/strings", http.StatusFound)
	}
}

func AdminStringsExport(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		activeProject := GetActiveProjectFromContext(r.Context())
		if activeProject == nil {
			http.Error(w, "No active project", http.StatusBadRequest)
			return
		}

		data, err := db.GetAllStringsForExport(deps.DB, activeProject.ID)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			http.Error(w, "JSON encoding error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_strings_export.json\"", activeProject.Name))
		w.Write(jsonBytes)
	}
}
