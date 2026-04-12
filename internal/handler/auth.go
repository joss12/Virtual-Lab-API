package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func Register(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if req.Email == "" || len(req.Password) < 8 {
			writeError(w, http.StatusBadRequest, "email required and password must be at least 8 characters")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		user, err := q.CreateUser(r.Context(), db.CreateUserParams{
			Email:        req.Email,
			PasswordHash: string(hash),
		})
		if err != nil {
			log.Printf("CreateUser error: %v", err)
			writeError(w, http.StatusConflict, "email already registered")
			return
		}

		token, err := middleware.IssueToken(user.ID.String())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not issue token")
			return
		}

		writeJSON(w, http.StatusCreated, authResponse{Token: token})
	}
}

func Login(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		user, err := q.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := middleware.IssueToken(user.ID.String())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not issue token")
			return
		}

		writeJSON(w, http.StatusOK, authResponse{Token: token})
	}
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func ChangePassword(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		var req changePasswordRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if len(req.NewPassword) < 8 {
			writeError(w, http.StatusBadRequest, "new password must be at least 8 characters")
			return
		}

		user, err := q.GetUserByID(r.Context(), parseUUID(userID))
		if err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
			writeError(w, http.StatusUnauthorized, "current password is incorrect")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if err := q.UpdateUserPassword(r.Context(), parseUUID(userID), string(hash)); err != nil {
			log.Printf("UpdateUserPassword error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not update password")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
	}
}
