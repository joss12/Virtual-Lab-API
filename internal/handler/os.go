package handler

import (
	"log"
	"net/http"

	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

func GetOsCourses(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courses, err := q.GetOsCourses(r.Context())
		if err != nil {
			log.Printf("GetOsCourses error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch courses")
			return
		}
		writeJSON(w, http.StatusOK, courses)
	}
}

func GetOsLessons(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseSlug := r.PathValue("course")
		if courseSlug == "" {
			writeError(w, http.StatusBadRequest, "course slug required")
			return
		}
		lessons, err := q.GetOsLessonsByCourse(r.Context(), courseSlug)
		if err != nil {
			log.Printf("GetOsLessons error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch lessons")
			return
		}
		writeJSON(w, http.StatusOK, lessons)
	}
}

func GetOsLesson(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lessonSlug := r.PathValue("lesson")
		if lessonSlug == "" {
			writeError(w, http.StatusBadRequest, "lesson slug required")
			return
		}
		lesson, err := q.GetOsLesson(r.Context(), lessonSlug)
		if err != nil {
			log.Printf("GetOsLesson error: %v", err)
			writeError(w, http.StatusNotFound, "lesson not found")
			return
		}
		writeJSON(w, http.StatusOK, lesson)
	}
}

func GetOsQuiz(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lessonSlug := r.PathValue("lesson")
		if lessonSlug == "" {
			writeError(w, http.StatusBadRequest, "lesson slug required")
			return
		}
		questions, err := q.GetOsQuizQuestions(r.Context(), lessonSlug)
		if err != nil {
			log.Printf("GetOsQuiz error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch quiz")
			return
		}
		writeJSON(w, http.StatusOK, questions)
	}
}

func GetOsProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)
		progress, err := q.GetOsProgress(r.Context(), parseUUID(userID))
		if err != nil {
			log.Printf("GetOsProgress error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not fetch progress")
			return
		}
		writeJSON(w, http.StatusOK, progress)
	}
}

type osProgressRequest struct {
	LessonSlug string `json:"lesson_slug"`
	Completed  bool   `json:"completed"`
	QuizScore  *int   `json:"quiz_score"`
	QuizTotal  *int   `json:"quiz_total"`
}

func UpdateOsProgress(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		var req osProgressRequest
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

		if err := q.UpsertOsProgress(r.Context(), parseUUID(userID), req.LessonSlug, req.Completed, req.QuizScore, req.QuizTotal); err != nil {
			log.Printf("UpdateOsProgress error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not update progress")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "progress updated"})
	}
}
