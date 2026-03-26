package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func StringsList(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeProject := GetActiveProjectFromContext(r.Context())

		// If no query params at all, restore from cookie
		if r.URL.RawQuery == "" {
			if c, err := r.Cookie("strings_filter"); err == nil && c.Value != "" {
				restored, err := url.QueryUnescape(c.Value)
				if err == nil && restored != "" {
					http.Redirect(w, r, "/strings?"+restored, http.StatusFound)
					return
				}
			}
		} else {
			// Save current query params to cookie using url.Values for safe encoding
			http.SetCookie(w, &http.Cookie{
				Name:    "strings_filter",
				Value:   url.QueryEscape(r.URL.RawQuery),
				Path:    "/",
				Expires: time.Now().Add(30 * 24 * time.Hour),
			})
		}

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
		sourceQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		targetQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("qt")))
		status := strings.ToLower(r.URL.Query().Get("status"))
		sortParam := r.URL.Query().Get("sort")
		idFilter := strings.TrimSpace(r.URL.Query().Get("id_filter"))

		totalItems, err := db.CountStrings(deps.DB, activeProject.ID, sourceQuery, targetQuery, status, idFilter)
		if err != nil {
			http.Error(w, "DB error counting", http.StatusInternalServerError)
			return
		}

		strs, err := db.GetStringsPaginated(deps.DB, activeProject.ID, sourceQuery, targetQuery, status, sortParam, idFilter, perPage, offset)
		if err != nil {
			http.Error(w, "DB error fetching", http.StatusInternalServerError)
			return
		}

		totalPages := (totalItems + perPage - 1) / perPage

		render.HTML(w, r, http.StatusOK, "strings_list.html", map[string]interface{}{
			"Strings":       strs,
			"Query":         sourceQuery,
			"QueryTarget":   targetQuery,
			"Status":        status,
			"Sort":          sortParam,
			"IDFilter":      idFilter,
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
				locale = "target" // Default locale
			}
			newText := r.FormValue("translation")

			err = db.UpsertTranslation(deps.DB, stringID, locale, newText, false, user.ID)
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
			// we only show revisions for the first locale
			revisions, _ = db.GetRevisionsForTranslation(deps.DB, translations[0].ID)
		}

		var filterQuery string
		if c, err := r.Cookie("strings_filter"); err == nil && c.Value != "" {
			if restored, err := url.QueryUnescape(c.Value); err == nil {
				filterQuery = restored
			}
		}

		// Fetch users for rendering avatars
		users, _ := db.GetAllUsers(deps.DB)
		usersMap := make(map[int]db.User)
		for _, u := range users {
			usersMap[u.ID] = u
		}

		render.HTML(w, r, http.StatusOK, "string_detail.html", map[string]interface{}{
			"String":        str,
			"Translations":  translations,
			"Revisions":     revisions,
			"Comments":      comments,
			"FilterQuery":   filterQuery,
			"Me":            user,
			"ActiveProject": activeProject,
			"UserProjects":  GetUserProjectsFromContext(r.Context()),
			"UsersMap":      usersMap,
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
				err = db.UpsertTranslation(deps.DB, stringID, "target", targetText, true, user.ID)
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

		filename := activeProject.ExportFilename
		if filename == "" {
			filename = activeProject.Name
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", filename))
		w.Write(jsonBytes)
	}
}

func AdminStringsExportPO(deps *app.App) http.HandlerFunc {
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

		stringsData, err := db.GetAllStringsWithTranslationsForExportPO(deps.DB, activeProject.ID)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filename := activeProject.ExportFilename
		if filename == "" {
			filename = activeProject.Name
		}

		w.Header().Set("Content-Type", "text/x-gettext-translation; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.po\"", filename))

		var sb strings.Builder
		sb.WriteString("msgid \"\"\n")
		sb.WriteString("msgstr \"\"\n")

		projIdVer := activeProject.PoProjectIdVersion
		if projIdVer == "" {
			projIdVer = activeProject.Name
		}
		sb.WriteString("\"Project-Id-Version: " + projIdVer + "\\n\"\n")

		if activeProject.PoReportMsgidBugsTo != "" {
			sb.WriteString(fmt.Sprintf("\"Report-Msgid-Bugs-To: %s\\n\"\n", activeProject.PoReportMsgidBugsTo))
		}

		if activeProject.PoLastTranslator != "" {
			sb.WriteString(fmt.Sprintf("\"Last-Translator: %s\\n\"\n", activeProject.PoLastTranslator))
		}

		if activeProject.PoLanguageTeam != "" {
			sb.WriteString(fmt.Sprintf("\"Language-Team: %s\\n\"\n", activeProject.PoLanguageTeam))
		}

		if activeProject.PoLanguage != "" {
			sb.WriteString(fmt.Sprintf("\"Language: %s\\n\"\n", activeProject.PoLanguage))
		}

		sb.WriteString("\"Content-Type: text/plain; charset=UTF-8\\n\"\n")
		sb.WriteString("\"Content-Transfer-Encoding: 8bit\\n\"\n\n")

		for _, s := range stringsData {
			if s.Context != nil && *s.Context != "" {
				sb.WriteString(fmt.Sprintf("msgctxt %q\n", *s.Context))
			}
			sb.WriteString(fmt.Sprintf("msgid %q\n", s.SourceText))

			if s.TargetTranslationText != nil && *s.TargetTranslationText != "" {
				sb.WriteString(fmt.Sprintf("msgstr %q\n", *s.TargetTranslationText))
			} else {
				sb.WriteString("msgstr \"\"\n")
			}
			sb.WriteString("\n")
		}

		w.Write([]byte(sb.String()))
	}
}

func AdminStringsImportPO(deps *app.App) http.HandlerFunc {
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

		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("po_file")
		if err != nil {
			http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var msgctxt, msgid, msgstr string
		var state string

		processEntry := func() {
			if msgid != "" {
				// Insert or update DB
				stringID, err := db.CreateString(deps.DB, activeProject.ID, "", msgid, msgctxt)
				if err == nil && msgstr != "" {
					_ = db.UpsertTranslation(deps.DB, stringID, "target", msgstr, true, user.ID)
				}
			}
			msgctxt, msgid, msgstr = "", "", ""
		}

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.HasPrefix(line, "msgctxt ") {
				state = "msgctxt"
				val, err := strconv.Unquote(strings.TrimSpace(line[8:]))
				if err == nil {
					msgctxt = val
				}
			} else if strings.HasPrefix(line, "msgid ") {
				if msgid != "" && state == "msgstr" {
					processEntry()
				}
				state = "msgid"
				val, err := strconv.Unquote(strings.TrimSpace(line[6:]))
				if err == nil {
					msgid = val
				}
			} else if strings.HasPrefix(line, "msgstr ") {
				state = "msgstr"
				val, err := strconv.Unquote(strings.TrimSpace(line[7:]))
				if err == nil {
					msgstr = val
				}
			} else if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
				val, err := strconv.Unquote(line)
				if err == nil {
					switch state {
					case "msgctxt":
						msgctxt += val
					case "msgid":
						msgid += val
					case "msgstr":
						msgstr += val
					}
				}
			}
		}

		// Process last entry
		processEntry()

		http.Redirect(w, r, "/strings", http.StatusFound)
	}
}
