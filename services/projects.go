package services

import (
	"app/urtc/db"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func NProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["owner"]
	userModel := &db.UserModel{
		DB: db.DB,
	}
	user, err := userModel.GetUser(username)
	if err != nil {
		fmt.Println("Invalid Username")
	} else {
		owner_id := user.ID
		projectModel := &db.ProjectModel{
			DB: db.DB,
		}
		projects, err := projectModel.GetProjectsByUser(owner_id)
		if err != nil {
			fmt.Println("Error : ", err)
		}
		fmt.Fprintf(w, "No of Projects : %d", len(projects))
	}
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["owner"]
	userModel := &db.UserModel{
		DB: db.DB,
	}
	user, err := userModel.GetUser(username)
	if err != nil {
		fmt.Println("Invalid Username")
	} else {
		owner_id := user.ID
		projectModel := &db.ProjectModel{
			DB: db.DB,
		}
		projects, err := projectModel.GetProjectsByUser(owner_id)
		if err != nil {
			fmt.Println("Error : ", err)
		}

		for i := 0; i < len(projects); i++ {
			project := projects[i]
			fmt.Fprintf(w, "ID : %d , Owner ID : %d, Name : %s, Description : %s, Created At : %s \n", project.ID, project.OwnerID, project.Name, project.Description, project.CreatedAt)
		}
	}
}
func GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["owner"]
	nameStr := vars["name"]

	userModel := &db.UserModel{
		DB: db.DB,
	}
	user, err := userModel.GetUser(username)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	} else {
		ownerID := user.ID
		projectModel := &db.ProjectModel{
			DB: db.DB,
		}
		project, err := projectModel.GetProjectByName(ownerID, nameStr)
		if err != nil {
			fmt.Println("Error : ", err)
		}

		fmt.Fprintf(w, "ID : %d , Owner ID : %d, Name : %s, Description : %s, Created At : %s \n", project.ID, project.OwnerID, project.Name, project.Description, project.CreatedAt)
	}
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["owner"]
	projectName := vars["name"]

	userModel := &db.UserModel{
		DB: db.DB,
	}
	user, err := userModel.GetUser(username)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	} else {
		ownerID := user.ID
		projectModel := &db.ProjectModel{
			DB: db.DB,
		}
		project, err := projectModel.GetProjectByName(ownerID, projectName)
		if err != nil {
			fmt.Println("Error : ", err)
		}
		projectID := project.ID
		err = projectModel.DeleteProject(ownerID, projectID)
		if err != nil {
			fmt.Println("Error : ", err)
		}
	}
}
