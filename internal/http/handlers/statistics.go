package handlers

import (
	"log"
	"net/http"
	"pierakladnia/internal/app"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func ProjectStatistics(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeProject := GetActiveProjectFromContext(r.Context())
		if activeProject == nil {
			http.Redirect(w, r, "/projects/switch", http.StatusFound)
			return
		}

		user := GetUserFromContext(r.Context())

		stats, total, err := db.GetProjectStatistics(deps.DB, activeProject.ID)
		if err != nil {
			log.Printf("Error fetching statistics: %v", err)
			http.Error(w, "Failed to load statistics", http.StatusInternalServerError)
			return
		}

		percentages := make(map[string]int)
		if total > 0 {
			for status, count := range stats {
				percentages[status] = int((float64(count) / float64(total)) * 100)
			}
		}

		// Ensure we don't have nil for statuses we expect
		for _, s := range []string{"todo", "draft", "needs_review", "done", "untranslated"} {
			if _, ok := stats[s]; !ok {
				stats[s] = 0
				percentages[s] = 0
			}
		}

		data := map[string]interface{}{
			"Me":            user,
			"ActiveProject": activeProject,
			"UserProjects":  GetUserProjectsFromContext(r.Context()),
			"Stats":         stats,
			"Percentages":   percentages,
			"TotalCount":    total,
		}

		render.HTML(w, http.StatusOK, "statistics.html", data)
	}
}
