// Package handler
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/middleware"
)

type githubUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Login string `json:"login"`
}

func getGithubEmails(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found")
}

func GithubLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		callbackURL := os.Getenv("GITHUB_CALLBACK_URL")
		url := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
			clientID, callbackURL,
		)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func GithubCallback(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		frontendURL := os.Getenv("FRONTEND_URL")
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, frontendURL+"/auth/login?error=no_code", http.StatusTemporaryRedirect)
			return
		}

		// Exchange code for access token
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

		tokenReqBody := strings.NewReader(fmt.Sprintf(
			`{"client_id":"%s","client_secret":"%s","code":"%s"}`,
			clientID, clientSecret, code,
		))
		req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", tokenReqBody)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("GitHub token exchange error: %v", err)
			http.Redirect(w, r, frontendURL+"/auth/login?error=token_exchange", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		var tokenResp struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		body, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub token response: %s", string(body))
		if err := json.Unmarshal(body, &tokenResp); err != nil || tokenResp.AccessToken == "" {
			log.Printf("GitHub token decode error: %v, body: %s", err, string(body))
			http.Redirect(w, r, frontendURL+"/auth/login?error=token_decode", http.StatusTemporaryRedirect)
			return
		}

		// Get GitHub user info
		userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		userReq.Header.Set("Accept", "application/vnd.github+json")
		userResp, err := http.DefaultClient.Do(userReq)
		if err != nil {
			http.Redirect(w, r, frontendURL+"/auth/login?error=user_fetch", http.StatusTemporaryRedirect)
			return
		}
		defer userResp.Body.Close()
		userBody, _ := io.ReadAll(userResp.Body)
		var ghUser githubUser
		if err := json.Unmarshal(userBody, &ghUser); err != nil {
			http.Redirect(w, r, frontendURL+"/auth/login?error=user_decode", http.StatusTemporaryRedirect)
			return
		}

		// Get email if not public
		if ghUser.Email == "" {
			ghUser.Email, err = getGithubEmails(tokenResp.AccessToken)
			if err != nil {
				http.Redirect(w, r, frontendURL+"/auth/login?error=no_email", http.StatusTemporaryRedirect)
				return
			}
		}
		ghUser.Email = strings.ToLower(strings.TrimSpace(ghUser.Email))
		githubID := fmt.Sprintf("%d", ghUser.ID)

		// Find or create user
		user, err := q.GetUserByGithubID(r.Context(), githubID)
		if err != nil {
			// Not found by github_id — try by email
			user, err = q.GetUserByEmail(r.Context(), ghUser.Email)
			if err != nil {
				// Create new user
				user, err = q.CreateGithubUser(context.Background(), db.CreateGithubUserParams{
					Email:    ghUser.Email,
					GithubID: githubID,
				})
				if err != nil {
					log.Printf("CreateGithubUser error: %v", err)
					http.Redirect(w, r, frontendURL+"/auth/login?error=create_user", http.StatusTemporaryRedirect)
					return
				}
			}
		}

		// Issue JWT
		token, err := middleware.IssueToken(user.ID.String())
		if err != nil {
			http.Redirect(w, r, frontendURL+"/auth/login?error=token_issue", http.StatusTemporaryRedirect)
			return
		}

		// Redirect to frontend with token
		http.Redirect(w, r, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, token), http.StatusTemporaryRedirect)
	}
}
