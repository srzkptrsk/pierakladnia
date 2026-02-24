package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
)

func CreateComment(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user := GetUserFromContext(r.Context())
		r.ParseForm()

		entityType := r.FormValue("entity_type")
		entityID, _ := strconv.Atoi(r.FormValue("entity_id"))
		content := strings.TrimSpace(r.FormValue("content"))
		parentIDStr := r.FormValue("parent_id")

		var parentID *int
		if parentIDStr != "" {
			pid, _ := strconv.Atoi(parentIDStr)
			parentID = &pid
		}

		if content == "" {
			http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
			return
		}

		if err := db.CreateComment(deps.DB, entityType, entityID, parentID, user.ID, content); err != nil {
			http.Error(w, "Failed to save comment", http.StatusInternalServerError)
			return
		}

		// Redirect back based on type
		if entityType == "translation" {
			http.Redirect(w, r, fmt.Sprintf("/strings/%d", entityID), http.StatusFound)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/glossary/%d", entityID), http.StatusFound)
		}
	}
}
