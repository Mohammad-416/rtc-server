package routers

import (
	"app/urtc/services"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Push Project
	r.HandleFunc("/push/manual", services.PushProject).Methods("POST")

	// Github Token Access
	r.HandleFunc("/db/token/{super_user_key}/{user}", services.GetToken).Methods("GET")

	// Project Functions
	r.HandleFunc("/db/projects-count/{owner}", services.NProjects).Methods("GET")
	r.HandleFunc("/db/projects/{owner}", services.GetProjects).Methods("GET")
	r.HandleFunc("/db/projects/{owner}/{name}", services.GetProject).Methods("GET")
	r.HandleFunc("/db/projects/{owner}/{name}", services.DeleteProject).Methods("DELETE")

	// User Functions
	r.HandleFunc("/db/users-count", services.GetUsersLen).Methods("GET")
	r.HandleFunc("/db/users", services.GetUsers).Methods("GET")
	r.HandleFunc("/db/users/{user}", services.GetUser).Methods("GET")
	r.HandleFunc("/db/user/{email}", services.GetUserByEmail).Methods("GET")
	r.HandleFunc("/db/users/{user}", services.DeleteUser).Methods("DELETE")

	// GitHub OAuth
	r.HandleFunc("/github/login", services.GitHubLoginHandler)
	r.HandleFunc("/github/callback", services.GitHubCallbackHandler)

	// Collaborator Routes
	r.HandleFunc("/collab/request", services.RequestCollaboration).Methods("POST")
	r.HandleFunc("/collab/approve", services.ApproveCollaboration).Methods("POST")
	r.HandleFunc("/collab/project", services.GetProjectCollaborators).Methods("GET")
	r.HandleFunc("/collab/user/requests", services.GetUserCollaborationRequests).Methods("GET")
	r.HandleFunc("/collab/token/{super_user_key}/{username}", services.GetCollaboratorToken).Methods("GET")
	r.HandleFunc("/collab/remove/{collab_id}", services.RemoveCollaborator).Methods("DELETE")

	return r
}
