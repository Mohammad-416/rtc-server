package services

import (
	"app/urtc/db"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileShareRequest struct {
	SenderEmail    string `json:"sender_email"`
	RecipientEmail string `json:"recipient_email"`
	ProjectID      string `json:"project_id,omitempty"`
	FileName       string `json:"file_name"`
	FileContent    string `json:"file_content"`
	FileType       string `json:"file_type"` // "code", "asset", "config", "scene", "prefab", "material", "shader", etc.
	Message        string `json:"message,omitempty"`
}

type CodeShareRequest struct {
	SenderEmail    string `json:"sender_email"`
	RecipientEmail string `json:"recipient_email"`
	CodeSnippet    string `json:"code_snippet"`
	Language       string `json:"language"`
	FileName       string `json:"file_name"`
	Message        string `json:"message,omitempty"`
}

type BulkFileShareRequest struct {
	SenderEmail    string     `json:"sender_email"`
	RecipientEmail string     `json:"recipient_email"`
	ProjectID      string     `json:"project_id,omitempty"`
	Files          []FileInfo `json:"files"`
	Message        string     `json:"message,omitempty"`
}

type FileInfo struct {
	FileName    string `json:"file_name"`
	FileContent string `json:"file_content"`
	FileType    string `json:"file_type"`
}

// Share file with another user in real-time
func ShareFile(w http.ResponseWriter, r *http.Request) {
	var req FileShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.SenderEmail == "" || req.RecipientEmail == "" || req.FileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "sender_email, recipient_email, and file_name are required",
		})
		return
	}

	// Validate sender
	userModel := &db.UserModel{DB: db.DB}
	sender, err := userModel.GetUserByEmail(req.SenderEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Sender not found",
		})
		return
	}

	// Validate recipient
	recipient, err := userModel.GetUserByEmail(req.RecipientEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Recipient not found",
		})
		return
	}

	// Prevent self-sharing
	if sender.ID == recipient.ID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot share files with yourself",
		})
		return
	}

	var projectName string

	// Check if recipient is an approved collaborator (if project_id is provided)
	if req.ProjectID != "" {
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid project ID",
			})
			return
		}

		// Verify project ownership
		projectModel := &db.ProjectModel{DB: db.DB}
		project, err := projectModel.GetProjectByID(projectID)
		if err != nil || project.OwnerID != sender.ID {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "You don't have permission to share from this project",
			})
			return
		}

		projectName = project.Name

		// Check if recipient is an approved collaborator
		collabModel := &db.CollaboratorModel{DB: db.DB}
		collaborators, _ := collabModel.GetApprovedCollaborators(projectID)

		isCollaborator := false
		for _, collab := range collaborators {
			if collab.Email == recipient.EMAIL {
				isCollaborator = true
				break
			}
		}

		if !isCollaborator {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Recipient is not an approved collaborator on this project",
			})
			return
		}
	}

	// Send file via WebSocket
	msg := Message{
		Type:           "file_share",
		SenderID:       sender.ID.String(),
		SenderEmail:    sender.EMAIL,
		RecipientID:    recipient.ID.String(),
		RecipientEmail: recipient.EMAIL,
		ProjectID:      req.ProjectID,
		ProjectName:    projectName,
		FileName:       req.FileName,
		FileContent:    req.FileContent,
		FileType:       req.FileType,
		Message:        req.Message,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	manager.broadcast <- msg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "File shared successfully",
		"recipient":    recipient.USERNAME,
		"is_online":    manager.isUserOnline(recipient.ID.String()),
		"file_name":    req.FileName,
		"project_name": projectName,
	})
}

// Share code snippet with another user
func ShareCode(w http.ResponseWriter, r *http.Request) {
	var req CodeShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.SenderEmail == "" || req.RecipientEmail == "" || req.CodeSnippet == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "sender_email, recipient_email, and code_snippet are required",
		})
		return
	}

	// Validate sender
	userModel := &db.UserModel{DB: db.DB}
	sender, err := userModel.GetUserByEmail(req.SenderEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Sender not found",
		})
		return
	}

	// Validate recipient
	recipient, err := userModel.GetUserByEmail(req.RecipientEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Recipient not found",
		})
		return
	}

	// Prevent self-sharing
	if sender.ID == recipient.ID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot share code with yourself",
		})
		return
	}

	// Send code via WebSocket
	metadata := map[string]interface{}{
		"language": req.Language,
		"filename": req.FileName,
	}

	msg := Message{
		Type:           "code_share",
		SenderID:       sender.ID.String(),
		SenderEmail:    sender.EMAIL,
		RecipientID:    recipient.ID.String(),
		RecipientEmail: recipient.EMAIL,
		FileContent:    req.CodeSnippet,
		Message:        req.Message,
		Timestamp:      time.Now().Format(time.RFC3339),
		Metadata:       metadata,
	}

	manager.broadcast <- msg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Code shared successfully",
		"recipient": recipient.USERNAME,
		"is_online": manager.isUserOnline(recipient.ID.String()),
		"language":  req.Language,
	})
}

// Share multiple files at once (bulk share)
func ShareBulkFiles(w http.ResponseWriter, r *http.Request) {
	var req BulkFileShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.SenderEmail == "" || req.RecipientEmail == "" || len(req.Files) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "sender_email, recipient_email, and files are required",
		})
		return
	}

	// Validate sender
	userModel := &db.UserModel{DB: db.DB}
	sender, err := userModel.GetUserByEmail(req.SenderEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Sender not found",
		})
		return
	}

	// Validate recipient
	recipient, err := userModel.GetUserByEmail(req.RecipientEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Recipient not found",
		})
		return
	}

	// Prevent self-sharing
	if sender.ID == recipient.ID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot share files with yourself",
		})
		return
	}

	var projectName string

	// Check collaboration if project_id is provided
	if req.ProjectID != "" {
		projectID, err := uuid.Parse(req.ProjectID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid project ID",
			})
			return
		}

		projectModel := &db.ProjectModel{DB: db.DB}
		project, err := projectModel.GetProjectByID(projectID)
		if err != nil || project.OwnerID != sender.ID {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "You don't have permission to share from this project",
			})
			return
		}

		projectName = project.Name

		collabModel := &db.CollaboratorModel{DB: db.DB}
		collaborators, _ := collabModel.GetApprovedCollaborators(projectID)

		isCollaborator := false
		for _, collab := range collaborators {
			if collab.Email == recipient.EMAIL {
				isCollaborator = true
				break
			}
		}

		if !isCollaborator {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Recipient is not an approved collaborator on this project",
			})
			return
		}
	}

	// Send bulk files as a single message
	fileNames := make([]string, len(req.Files))
	for i, file := range req.Files {
		fileNames[i] = file.FileName
	}

	msg := Message{
		Type:           "bulk_file_share",
		SenderID:       sender.ID.String(),
		SenderEmail:    sender.EMAIL,
		RecipientID:    recipient.ID.String(),
		RecipientEmail: recipient.EMAIL,
		ProjectID:      req.ProjectID,
		ProjectName:    projectName,
		Message:        req.Message,
		Timestamp:      time.Now().Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"files":      req.Files,
			"file_count": len(req.Files),
		},
	}

	manager.broadcast <- msg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Files shared successfully",
		"recipient":    recipient.USERNAME,
		"is_online":    manager.isUserOnline(recipient.ID.String()),
		"file_count":   len(req.Files),
		"file_names":   strings.Join(fileNames, ", "),
		"project_name": projectName,
	})
}

// Get approved collaborators for file sharing with online status
func GetShareableCollaborators(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "project_id is required",
		})
		return
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid project ID",
		})
		return
	}

	collabModel := &db.CollaboratorModel{DB: db.DB}
	collaborators, err := collabModel.GetApprovedCollaborators(projectUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch collaborators",
		})
		return
	}

	// Add online status to each collaborator
	type CollaboratorWithStatus struct {
		db.CollaboratorWithDetails
		IsOnline bool `json:"is_online"`
	}

	result := make([]CollaboratorWithStatus, 0, len(collaborators))
	for _, collab := range collaborators {
		// Get user ID from email
		userModel := &db.UserModel{DB: db.DB}
		user, err := userModel.GetUserByEmail(collab.Email)
		if err != nil {
			continue
		}

		result = append(result, CollaboratorWithStatus{
			CollaboratorWithDetails: collab,
			IsOnline:                manager.isUserOnline(user.ID.String()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"project_id":    projectID,
		"collaborators": result,
		"total":         len(result),
		"online_count":  countOnlineCollaborators(result),
	})
}

// Helper function to count online collaborators
func countOnlineCollaborators(collaborators []struct {
	db.CollaboratorWithDetails
	IsOnline bool `json:"is_online"`
}) int {
	count := 0
	for _, c := range collaborators {
		if c.IsOnline {
			count++
		}
	}
	return count
}
