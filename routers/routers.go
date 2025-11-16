package routers

import (
	"app/urtc/services"
	"net/http"

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

	// WebSocket Routes
	r.HandleFunc("/ws", services.HandleWebSocket).Methods("GET")
	r.HandleFunc("/ws/online-users", services.GetOnlineUsers).Methods("GET")
	r.HandleFunc("/ws/user-status", services.CheckUserOnlineStatus).Methods("GET")

	// File Sharing Routes
	r.HandleFunc("/share/file", services.ShareFile).Methods("POST")
	r.HandleFunc("/share/code", services.ShareCode).Methods("POST")
	r.HandleFunc("/share/bulk", services.ShareBulkFiles).Methods("POST")
	r.HandleFunc("/share/collaborators", services.GetShareableCollaborators).Methods("GET")

	// Activity Tracking Routes
	r.HandleFunc("/activity/user", services.GetUserActivities).Methods("GET")
	r.HandleFunc("/activity/project", services.GetProjectActivities).Methods("GET")
	r.HandleFunc("/activity/team", services.GetRecentTeamActivities).Methods("GET")

	// Version Control Routes
	r.HandleFunc("/version/commit", services.CommitFileVersion).Methods("POST")
	r.HandleFunc("/version/history", services.GetFileHistory).Methods("GET")
	r.HandleFunc("/version/project", services.GetProjectVersions).Methods("GET")
	r.HandleFunc("/version/conflicts", services.GetFileConflicts).Methods("GET")
	r.HandleFunc("/version/resolve", services.ResolveConflict).Methods("POST")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return r
}
