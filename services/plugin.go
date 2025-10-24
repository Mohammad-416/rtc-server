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

type CollabRequest struct {
	CollaboratorEmail string `json:"collaborator_email"`
	ProjectID         string `json:"project_id"`
}

type CollabApproval struct {
	CollabID string `json:"collab_id"`
	Status   string `json:"status"`
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

func RequestCollaboration(w http.ResponseWriter, r *http.Request) {
	var req CollabRequest
	json.NewDecoder(r.Body).Decode(&req)

	userModel := &db.UserModel{DB: db.DB}
	user, err := userModel.GetUserByEmail(req.CollaboratorEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "User not found")
		return
	}

	query := `INSERT INTO collaborators (user_id, project_id) VALUES ($1, $2) RETURNING id`
	var collabID string
	err = db.DB.QueryRow(query, user.ID, req.ProjectID).Scan(&collabID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error creating request")
		return
	}

	fmt.Fprintf(w, "Collaboration request created: %s", collabID)
}

func ApproveCollaboration(w http.ResponseWriter, r *http.Request) {
	var req CollabApproval
	json.NewDecoder(r.Body).Decode(&req)

	query := `UPDATE collaborators SET status = $1 WHERE id = $2`
	_, err := db.DB.Exec(query, req.Status, req.CollabID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error updating status")
		return
	}

	fmt.Fprintf(w, "Collaboration %s", req.Status)
}
