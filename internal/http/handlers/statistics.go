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
		activeProject := db.Project{}
		if p, ok := r.Context().Value("ActiveProject").(*db.Project); ok && p != nil {
			activeProject = *p
		} else {
			http.Redirect(w, r, "/projects/switch", http.StatusFound)
			return
		}

		user := r.Context().Value("Me").(*db.User)

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
		for _, s := range []string{"todo", "draft", "translated", "approved", "untranslated"} {
			if _, ok := stats[s]; !ok {
				stats[s] = 0
				percentages[s] = 0
			}
		}

		data := struct {
			Me            *db.User
			ActiveProject *db.Project
			Stats         map[string]int
			Percentages   map[string]int
			TotalCount    int
		}{
			Me:            user,
			ActiveProject: &activeProject,
			Stats:         stats,
			Percentages:   percentages,
			TotalCount:    total,
		}

		render.HTML(w, http.StatusOK, "statistics.html", data)
	}
}
