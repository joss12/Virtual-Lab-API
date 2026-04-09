package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

type updateProgressRequest struct {
	TabsVisited []string `json:"tabs_visited"`
	Completed   bool     `json:"completed"`
}

func GetProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		rows, err := q.GetUserProgress(r.Context(), parseUUID(userID))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not fetch progress")
			return
		}

		writeJSON(w, http.StatusOK, rows)
	}
}

func UpdateProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)
		component := chi.URLParam(r, "component")

		if component == "" {
			writeError(w, http.StatusBadRequest, "component is required")
			return
		}

		var req updateProgressRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		progress, err := q.UpsertProgress(r.Context(), db.UpsertProgressParams{
			UserID:      parseUUID(userID),
			Component:   component,
			TabsVisited: req.TabsVisited,
			Completed:   req.Completed,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not update progress")
			return
		}

		writeJSON(w, http.StatusOK, progress)
	}
}
