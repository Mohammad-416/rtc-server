package services

import (
	"app/urtc/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Option 1: Initialize config in init function
var githubOAuthConfig *oauth2.Config

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"repo", "user"},
		Endpoint:     github.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"),
	}
}

func GitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Config:", githubOAuthConfig)
	url := githubOAuthConfig.AuthCodeURL("randomstate")
	http.Redirect(w, r, url, http.StatusFound)
}

func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user struct {
		Login string `json:"login"` // GitHub username
		ID    int    `json:"id"`    // GitHub ID
		Email string `json:"email"` // Might be empty if private
		Name  string `json:"name"`  // Full name (optional)
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Step 2: If email is empty, fetch /user/emails
	if user.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get user emails", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}

		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			http.Error(w, "Failed to parse user emails", http.StatusInternalServerError)
			return
		}

		for _, e := range emails {
			if e.Primary && e.Verified {
				user.Email = e.Email
				break
			}
		}
	}

	// Step 3: Ensure we got the email
	if user.Email == "" {
		http.Error(w, "Email not available, please make your email public on GitHub", http.StatusBadRequest)
		return
	}

	// Step 4: Store in DB
	// Assuming you have access to your initialized UserModel here
	userModel := &db.UserModel{
		DB: db.DB,
	}
	newUser, err := userModel.CreateUser(int64(user.ID), user.Login, user.Email)
	if err != nil {
		http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Welcome, %s! Your email is %s", newUser.USERNAME, newUser.EMAIL)
}
