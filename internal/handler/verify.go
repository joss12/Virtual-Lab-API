package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vlab-api/internal/db"
)

func sendVerificationEmail(toEmail, verifyURL string) error {
	apiKey := os.Getenv("RESEND_API_KEY")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:monospace;max-width:480px;margin:0 auto;padding:32px;background:#0d1420;color:#ddeeff;">
  <h2 style="margin:0 0 16px;font-size:20px;">Verify your email</h2>
  <p style="color:rgba(180,210,240,.7);line-height:1.7;margin:0 0 24px;">
    Click the button below to verify your vlab account. This link expires in 24 hours.
  </p>
  <a href="%s" style="display:inline-block;padding:12px 24px;background:#2563eb;color:white;text-decoration:none;border-radius:6px;font-family:monospace;font-size:14px;">
    Verify Email
  </a>
  <p style="color:rgba(180,210,240,.4);font-size:11px;margin:24px 0 0;">
    If you did not create this account, ignore this email.
  </p>
</body>
</html>`, verifyURL)

	escapedHTML := strings.ReplaceAll(html, `"`, `\"`)
	escapedHTML = strings.ReplaceAll(escapedHTML, "\n", "\\n")

	payload := fmt.Sprintf(`{
		"from": "vlab <noreply@mail.eddymouity.dev>",
		"to": ["%s"],
		"subject": "Verify your vlab email",
		"html": "%s"
	}`, toEmail, escapedHTML)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("resend request error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend send error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func SendVerificationOnRegister(q *db.Queries, userID fmt.Stringer, email string) {
	go func() {
		ctx := context.Background()

		token, err := generateToken()
		if err != nil {
			log.Printf("generateToken error: %v", err)
			return
		}

		err = q.CreateEmailVerification(ctx, db.CreateEmailVerificationParams{
			UserID:    parseUUID(userID.String()),
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			log.Printf("CreateEmailVerification error: %v", err)
			return
		}

		frontendURL := os.Getenv("FRONTEND_URL")
		verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", frontendURL, token)

		if err := sendVerificationEmail(email, verifyURL); err != nil {
			log.Printf("sendVerificationEmail error: %v", err)
		}
	}()
}

func VerifyEmail(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeError(w, http.StatusBadRequest, "token is required")
			return
		}

		userID, err := q.GetEmailVerification(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid or expired verification token")
			return
		}

		if err := q.VerifyUserEmail(r.Context(), userID); err != nil {
			log.Printf("VerifyUserEmail error: %v", err)
			writeError(w, http.StatusInternalServerError, "could not verify email")
			return
		}

		if err := q.DeleteEmailVerification(r.Context(), token); err != nil {
			log.Printf("DeleteEmailVerification error: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
	}
}
