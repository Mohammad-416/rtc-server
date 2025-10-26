package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Collaborator struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	Status    string    `json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CollaboratorWithDetails struct {
	ID            uuid.UUID `json:"collab_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ProjectName   string    `json:"project_name,omitempty"`
	ProjectDesc   string    `json:"project_description,omitempty"`
	OwnerUsername string    `json:"owner_username,omitempty"`
}

type CollaboratorModel struct {
	DB *sql.DB
}

// CreateCollaboration - Creates a new collaboration request with default pending status
func (m *CollaboratorModel) CreateCollaboration(userID, projectID uuid.UUID, status string) (Collaborator, error) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	// Default to pending if no status provided
	if status == "" {
		status = "pending"
	}

	query := `
		INSERT INTO collaborators (id, user_id, project_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, project_id, status, created_at, updated_at
	`

	var collab Collaborator
	err := m.DB.QueryRow(query, id, userID, projectID, status, createdAt, updatedAt).
		Scan(&collab.ID, &collab.UserID, &collab.ProjectID, &collab.Status, &collab.CreatedAt, &collab.UpdatedAt)

	if err != nil {
		return Collaborator{}, err
	}

	return collab, nil
}

// GetCollaborationByID - Gets a specific collaboration by ID
func (m *CollaboratorModel) GetCollaborationByID(collabID uuid.UUID) (Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE id = $1
	`

	var collab Collaborator
	err := m.DB.QueryRow(query, collabID).
		Scan(&collab.ID, &collab.UserID, &collab.ProjectID, &collab.Status, &collab.CreatedAt, &collab.UpdatedAt)

	if err != nil {
		return Collaborator{}, err
	}

	return collab, nil
}

// UpdateCollaborationStatus - Updates collaboration status (approved/rejected)
func (m *CollaboratorModel) UpdateCollaborationStatus(collabID uuid.UUID, status string) error {
	query := `
		UPDATE collaborators 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	updatedAt := time.Now()
	_, err := m.DB.Exec(query, status, updatedAt, collabID)
	return err
}

// GetProjectCollaborators - Gets all collaborators for a project with user details
func (m *CollaboratorModel) GetProjectCollaborators(projectID uuid.UUID) ([]CollaboratorWithDetails, error) {
	query := `
		SELECT c.id, c.status, c.created_at, c.updated_at, u.username, u.email
		FROM collaborators c
		JOIN users u ON c.user_id = u.id
		WHERE c.project_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := m.DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []CollaboratorWithDetails
	for rows.Next() {
		var collab CollaboratorWithDetails
		err := rows.Scan(&collab.ID, &collab.Status, &collab.CreatedAt, &collab.UpdatedAt,
			&collab.Username, &collab.Email)
		if err != nil {
			continue
		}
		collaborators = append(collaborators, collab)
	}

	return collaborators, nil
}

// GetUserPendingRequests - Gets all pending collaboration requests for a user
func (m *CollaboratorModel) GetUserPendingRequests(userID uuid.UUID) ([]CollaboratorWithDetails, error) {
	query := `
		SELECT c.id, c.status, c.created_at, c.updated_at, 
		       p.name as project_name, p.description as project_desc, 
		       u.username as owner_username
		FROM collaborators c
		JOIN projects p ON c.project_id = p.id
		JOIN users u ON p.owner_id = u.id
		WHERE c.user_id = $1 AND c.status = 'pending'
		ORDER BY c.created_at DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []CollaboratorWithDetails
	for rows.Next() {
		var req CollaboratorWithDetails
		err := rows.Scan(&req.ID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&req.ProjectName, &req.ProjectDesc, &req.OwnerUsername)
		if err != nil {
			continue
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetUserAllCollaborations - Gets all collaborations for a user (any status)
func (m *CollaboratorModel) GetUserAllCollaborations(userID uuid.UUID) ([]CollaboratorWithDetails, error) {
	query := `
		SELECT c.id, c.status, c.created_at, c.updated_at, 
		       p.name as project_name, p.description as project_desc, 
		       u.username as owner_username
		FROM collaborators c
		JOIN projects p ON c.project_id = p.id
		JOIN users u ON p.owner_id = u.id
		WHERE c.user_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborations []CollaboratorWithDetails
	for rows.Next() {
		var collab CollaboratorWithDetails
		err := rows.Scan(&collab.ID, &collab.Status, &collab.CreatedAt, &collab.UpdatedAt,
			&collab.ProjectName, &collab.ProjectDesc, &collab.OwnerUsername)
		if err != nil {
			continue
		}
		collaborations = append(collaborations, collab)
	}

	return collaborations, nil
}

// CheckCollaborationExists - Checks if collaboration already exists
func (m *CollaboratorModel) CheckCollaborationExists(userID, projectID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) FROM collaborators 
		WHERE user_id = $1 AND project_id = $2
	`

	var count int
	err := m.DB.QueryRow(query, userID, projectID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteCollaboration - Removes a collaborator
func (m *CollaboratorModel) DeleteCollaboration(collabID uuid.UUID) error {
	query := `DELETE FROM collaborators WHERE id = $1`
	_, err := m.DB.Exec(query, collabID)
	return err
}

// GetApprovedCollaborators - Gets only approved collaborators for a project
func (m *CollaboratorModel) GetApprovedCollaborators(projectID uuid.UUID) ([]CollaboratorWithDetails, error) {
	query := `
		SELECT c.id, c.status, c.created_at, c.updated_at, u.username, u.email
		FROM collaborators c
		JOIN users u ON c.user_id = u.id
		WHERE c.project_id = $1 AND c.status = 'approved'
		ORDER BY c.created_at DESC
	`

	rows, err := m.DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []CollaboratorWithDetails
	for rows.Next() {
		var collab CollaboratorWithDetails
		err := rows.Scan(&collab.ID, &collab.Status, &collab.CreatedAt, &collab.UpdatedAt,
			&collab.Username, &collab.Email)
		if err != nil {
			continue
		}
		collaborators = append(collaborators, collab)
	}

	return collaborators, nil
}
