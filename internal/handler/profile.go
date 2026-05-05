package handler

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

func GravatarURL(email string) string {
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=200", hash)
}

type ProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

func GetMe(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		user, err := q.GetUserByID(r.Context(), parseUUID(userID))
		if err != nil {
			writeJSON(w, http.StatusNotFound, "user not found")
			return
		}

		writeJSON(w, http.StatusOK, ProfileResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			AvatarURL: GravatarURL(user.Email),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}
