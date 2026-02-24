package handlers

import (
	"fmt"
	"net/http"

	"pierakladnia/internal/app"
)

func Health(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.DB.Ping(); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "OK")
	}
}
