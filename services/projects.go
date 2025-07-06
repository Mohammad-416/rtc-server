package services

import (
	"app/urtc/db"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func NProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerIDStr := vars["owner"]
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}
	projectModel := &db.ProjectModel{
		DB: db.DB,
	}
	projects, err := projectModel.GetProjectsByUser(ownerID)
	if err != nil {
		fmt.Println("Error : ", err)
	}
	fmt.Fprintf(w, "No of Projects : %d", len(projects))
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerIDStr := vars["owner"]
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}
	projectModel := &db.ProjectModel{
		DB: db.DB,
	}
	projects, err := projectModel.GetProjectsByUser(ownerID)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	for i := 0; i < len(projects); i++ {
		project := projects[i]
		fmt.Fprintf(w, "ID : %d , Owner ID : %d, Name : %s, Description : %s, Created At : %s \n", project.ID, project.OwnerID, project.Name, project.Description, project.CreatedAt)
	}
}
func GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerIDStr := vars["owner"]
	nameStr := vars["name"]
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}
	projectModel := &db.ProjectModel{
		DB: db.DB,
	}
	project, err := projectModel.GetProjectByName(ownerID, nameStr)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	fmt.Fprintf(w, "ID : %d , Owner ID : %d, Name : %s, Description : %s, Created At : %s \n", project.ID, project.OwnerID, project.Name, project.Description, project.CreatedAt)

}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerIDStr := vars["owner"]
	ownerID, err := uuid.Parse(ownerIDStr)
	projectIDStr := vars["id"]
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}

	projectModel := &db.ProjectModel{
		DB: db.DB,
	}

	err = projectModel.DeleteProject(ownerID, projectID)
	if err != nil {
		fmt.Println("Error : ", err)
	}

}
