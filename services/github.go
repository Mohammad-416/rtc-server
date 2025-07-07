package services

import (
	"app/urtc/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Initialize config in init function
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

	accessToken := token.AccessToken
	fmt.Println(accessToken)

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

	//If email is empty, fetch /user/emails
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

	//Ensure we got the email
	if user.Email == "" {
		http.Error(w, "Email not available, please make your email public on GitHub", http.StatusBadRequest)
		return
	}

	// Store in DB
	userModel := &db.UserModel{
		DB: db.DB,
	}

	// Check if the user already exists
	newUser, err := userModel.GetUserByEmail(user.Email)
	if err == nil {
		//Store the github data
		stored_data := StoreAccessToken(newUser.USERNAME, accessToken, newUser.ID)
		if stored_data {
			fmt.Println("Data Updated Successfully")
			w.WriteHeader(http.StatusOK)
		} else {
			fmt.Println("Unable to update token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Welcome, %s! Your email is %s", newUser.USERNAME, newUser.EMAIL)
	} else {
		newUser, err := userModel.CreateUser(int64(user.ID), user.Login, user.Email)
		if err != nil {
			http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			//Store the github data
			stored_data := StoreAccessToken(newUser.USERNAME, accessToken, newUser.ID)
			if stored_data {
				fmt.Println("Data Stored Successfully")
				w.WriteHeader(http.StatusOK)
			} else {
				fmt.Println("Unable to save token")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		fmt.Fprintf(w, "Welcome, %s! Your email is %s", newUser.USERNAME, newUser.EMAIL)
	}

	fmt.Fprintf(w, "You can now continue working in your game engine")
}

func StoreAccessToken(username, githubToken string, user_id uuid.UUID) bool {

	tokenModel := &db.TokenModel{
		DB: db.DB,
	}

	github_data, err := tokenModel.GetToken(username)
	if err != nil {
		github_data, err := tokenModel.SaveToken(githubToken, username, user_id)
		if err != nil {
			fmt.Println("Error : ", err)
			fmt.Println("User created but unable to save github access token")
			return false
		} else {
			fmt.Println(github_data)
			fmt.Println("Saved the github token successfully")
			return true
		}
	} else {
		fmt.Println(github_data)
		fmt.Println("Data already exists, Trying to update the data")
		err := tokenModel.UpdateToken(username, githubToken)
		if err != nil {
			fmt.Println("Error : ", err)
			return false
		} else {
			fmt.Println("Updated Successfully")
			return true
		}
	}
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["user"]
	super_user_key := vars["super_user_key"]

	if super_user_key != os.Getenv("SECRET_KEY") {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	tokenModel := &db.TokenModel{
		DB: db.DB,
	}
	github_data, err := tokenModel.GetToken(username)
	if err != nil {
		fmt.Println("Error : ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "ID : %d , Username : %s, Github Token : %s, User ID : %d, Created At : %s \n", github_data.ID, github_data.USERNAME, github_data.GITHUB_TOKEN, github_data.USER_ID, github_data.CREATED_AT)

}
