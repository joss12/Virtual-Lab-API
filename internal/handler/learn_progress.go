// Package handler
package handler

import (
	"log"
	"net/http"

	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

func GetLearnProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)
		language := r.PathValue("language")
		if language == "" {
			writeError(w, http.StatusBadRequest, "language is required")
			return
		}

		progress, err := q.GetLearnProgress(r.Context(), parseUUID(userID), language)
		if err != nil {
			log.Printf("GetLearnProgress error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch progress")
			return
		}
		writeJSON(w, http.StatusOK, progress)
	}
}

type learnProgressRequest struct {
	LessonSlug string `json:"lesson_slug"`
	Completed  bool   `json:"completed"`
	QuizScore  *int   `json:"quiz_score"`
	QuizTotal  *int   `json:"quiz_total"`
}

func UpdateLearnProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)
		language := r.PathValue("language")
		if language == "" {
			writeError(w, http.StatusBadRequest, "language is required")
			return
		}

		var req learnProgressRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.LessonSlug == "" {
			writeError(w, http.StatusBadRequest, "lesson_slug is required")
			return
		}

		// Enforce 75% pass rate for quiz completion
		if req.QuizScore != nil && req.QuizTotal != nil && *req.QuizTotal > 0 {
			pct := (*req.QuizScore * 100) / *req.QuizTotal
			if pct < 75 {
				req.Completed = false
			}
		}

		if err := q.UpsertLearnProgress(r.Context(), parseUUID(userID), language, req.LessonSlug, req.Completed, req.QuizScore, req.QuizTotal); err != nil {
			log.Printf("UpdateLearnProgress error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not update progress")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "progress updated"})
	}
}

func GetLearnProgressSummary(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		summary, err := q.GetLearnProgressSummary(r.Context(), parseUUID(userID))
		if err != nil {
			log.Printf("GetLearnProgressSummary error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch summary")
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}
