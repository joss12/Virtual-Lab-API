package handler

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

type saveScoreRequest struct {
	Score     int    `json:"score"`
	Total     int    `json:"total"`
	Component string `json:"component"`
}

func SaveScore(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		var req saveScoreRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Total <= 0 || req.Score < 0 || req.Score > req.Total {
			writeError(w, http.StatusBadRequest, "invalid score values")
			return
		}

		score, err := q.InsertQuizScore(r.Context(), db.InsertQuizScoreParams{
			UserID:    parseUUID(userID),
			Score:     int32(req.Score),
			Total:     int32(req.Total),
			Component: pgtype.Text{String: req.Component, Valid: req.Component != ""},
		})
		if err != nil {
			log.Printf("InsertQuizScore error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not save score")
			return
		}

		writeJSON(w, http.StatusCreated, score)
	}
}

func GetScores(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		scores, err := q.GetUserScores(r.Context(), parseUUID(userID))
		if err != nil {
			log.Printf("GetUserScores error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch scores")
			return
		}

		writeJSON(w, http.StatusOK, scores)
	}
}
