package services

import (
	"app/urtc/db"
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

			fmt.Fprintln(w, "success : ", http.StatusOK)
			fmt.Fprintln(w, "message : Collaboration started successfully for project ", project.Name)
			fmt.Fprintln(w, "project_id : ", project.ID)

		} else {
			fmt.Println("Project already exists with the name : ", project.Name)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

}
