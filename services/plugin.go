package services

import (
	"app/urtc/db"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MetaUser struct {
	EMAIL        string `json:"user_email"`
	PROJECT_NAME string `json:"project_name"`
}

func PushProject(w http.ResponseWriter, r *http.Request) {

	var metaUser MetaUser
	json.NewDecoder(r.Body).Decode(&metaUser)

	userModel := &db.UserModel{
		DB: db.DB,
	}
	user, err := userModel.GetUserByEmail(metaUser.EMAIL)
	if err != nil {
		fmt.Println("Error : ", err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		fmt.Println(user.USERNAME, user.ID, user.GITHUB_ID, user.EMAIL, user.CREATED_AT)

		//starting a new project from here
		projectModel := &db.ProjectModel{
			DB: db.DB,
		}

		tokenModel := &db.TokenModel{
			DB: db.DB,
		}

		token, err := tokenModel.GetToken(user.USERNAME)
		if err != nil {
			fmt.Println("Error Fetching Token : ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//Check if a project by this name already exists
		project, err := projectModel.GetProjectByName(user.ID, metaUser.PROJECT_NAME)
		if err != nil {
			project, err := projectModel.CreateProject(user.ID, metaUser.PROJECT_NAME, metaUser.PROJECT_NAME)
			if err != nil {
				fmt.Println("Error : ", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Println("Project Created Succesfully")

			//Start Initializing a github repo here
			payload := map[string]interface{}{
				"name":    metaUser.PROJECT_NAME,
				"private": false,
			}

			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", "https://api.github.com/user/repos", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "token "+token.GITHUB_TOKEN)
			req.Header.Set("Accept", "application/vnd.github.v3+json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				fmt.Println("Repository created successfully!")
			} else {
				fmt.Println("Failed to create repository. Status:", resp.Status)
			}

			var repos struct {
				Name    string `json:"name"`
				HTMLURL string `json:"html_url"` // This is the repo link
			}

			if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
				http.Error(w, "Failed to parse repos", http.StatusInternalServerError)
				return
			}

			fmt.Fprintln(w, "success : ", http.StatusOK)
			fmt.Fprintln(w, "message : Collaboration started successfully for project ", project.Name)
			fmt.Fprintln(w, "project_id : ", project.ID)
			fmt.Fprintf(w, "github repo URL: %s", repos.HTMLURL)

		} else {
			fmt.Println("Project already exists with the name : ", project.Name)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

}
