package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/vlab-api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sendResetEmail(toEmail, resetURL string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	auth := smtp.PlainAuth("", user, pass, host)

	body := fmt.Sprintf(`From: vlab <%s>
To: %s
Subject: Reset your vlab password
MIME-Version: 1.0
Content-Type: text/html; charset="UTF-8"

<!DOCTYPE html>
<html>
<body style="font-family:monospace;max-width:480px;margin:0 auto;padding:32px;background:#0d1420;color:#ddeeff;">
  <h2 style="margin:0 0 16px;font-size:20px;">Reset your password</h2>
  <p style="color:rgba(180,210,240,.7);line-height:1.7;margin:0 0 24px;">
    Click the button below to reset your vlab password. This link expires in 1 hour.
  </p>
  <a href="%s" style="display:inline-block;padding:12px 24px;background:#2563eb;color:white;text-decoration:none;border-radius:6px;font-family:monospace;font-size:14px;">
    Reset Password
  </a>
  <p style="color:rgba(180,210,240,.4);font-size:11px;margin:24px 0 0;">
    If you did not request this, ignore this email. Your password will not change.
  </p>
</body>
</html>`, user, toEmail, resetURL)

	err := smtp.SendMail(
		host+":"+port,
		auth,
		user,
		[]string{toEmail},
		[]byte(body),
	)
	if err != nil {
		return fmt.Errorf("smtp error: %v", err)
	}
	return nil
}

type forgotRequest struct {
	Email string `json:"email"`
}

func ForgotPassword(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req forgotRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Always return success to prevent email enumeration
		writeJSON(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})

		// Do the actual work in background
		go func() {
			ctx := context.Background()
			user, err := q.GetUserByEmail(ctx, req.Email)
			if err != nil {
				return // User not found — silently do nothing
			}

			token, err := generateToken()
			if err != nil {
				log.Printf("generateToken error: %v", err)
				return
			}

			err = q.CreatePasswordReset(ctx, db.CreatePasswordResetParams{
				UserID:    user.ID,
				Token:     token,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			})
			if err != nil {
				log.Printf("CreatePasswordReset error: %v", err)
				return
			}

			frontendURL := os.Getenv("FRONTEND_URL")
			resetURL := fmt.Sprintf("%s/auth/reset?token=%s", frontendURL, token)

			if err := sendResetEmail(user.Email, resetURL); err != nil {
				log.Printf("sendResetEmail error: %v", err)
			}
		}()
	}
}

type resetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func ResetPassword(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req resetRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Token == "" {
			writeError(w, http.StatusBadRequest, "token is required")
			return
		}

		if len(req.NewPassword) < 8 {
			writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}

		reset, err := q.GetPasswordReset(r.Context(), req.Token)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid or expired reset token")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if err := q.UpdateUserPassword(r.Context(), reset.UserID, string(hash)); err != nil {
			log.Printf("UpdateUserPassword error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not update password")
			return
		}

		if err := q.DeletePasswordReset(r.Context(), req.Token); err != nil {
			log.Printf("DeletePasswordReset error: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
	}
}
