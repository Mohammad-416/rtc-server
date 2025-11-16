package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Collaborator struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProjectID uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CollaboratorModel struct {
	DB *sql.DB
}

// CreateCollaboration - Creates a new collaboration request
func (m *CollaboratorModel) CreateCollaboration(userID uuid.UUID, projectID uuid.UUID, status string) (*Collaborator, error) {
	query := `
		INSERT INTO collaborators (id, user_id, project_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, project_id, status, created_at, updated_at
	`

	id := uuid.New()
	now := time.Now()

	var collab Collaborator
	err := m.DB.QueryRow(query, id, userID, projectID, status, now, now).Scan(
		&collab.ID,
		&collab.UserID,
		&collab.ProjectID,
		&collab.Status,
		&collab.CreatedAt,
		&collab.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// GetCollaborationByID - Retrieves collaboration by ID
func (m *CollaboratorModel) GetCollaborationByID(collabID uuid.UUID) (*Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE id = $1
	`

	var collab Collaborator
	err := m.DB.QueryRow(query, collabID).Scan(
		&collab.ID,
		&collab.UserID,
		&collab.ProjectID,
		&collab.Status,
		&collab.CreatedAt,
		&collab.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// GetCollaborationByUserAndProject - Get collaboration by user and project
func (m *CollaboratorModel) GetCollaborationByUserAndProject(userID uuid.UUID, projectID uuid.UUID) (*Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE user_id = $1 AND project_id = $2
	`

	var collab Collaborator
	err := m.DB.QueryRow(query, userID, projectID).Scan(
		&collab.ID,
		&collab.UserID,
		&collab.ProjectID,
		&collab.Status,
		&collab.CreatedAt,
		&collab.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// UpdateCollaborationStatus - Updates collaboration status
func (m *CollaboratorModel) UpdateCollaborationStatus(collabID uuid.UUID, status string) error {
	query := `
		UPDATE collaborators
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := m.DB.Exec(query, status, now, collabID)
	return err
}

// GetProjectCollaborators - Gets all collaborators for a project
func (m *CollaboratorModel) GetProjectCollaborators(projectID uuid.UUID) ([]Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []Collaborator
	for rows.Next() {
		var collab Collaborator
		err := rows.Scan(
			&collab.ID,
			&collab.UserID,
			&collab.ProjectID,
			&collab.Status,
			&collab.CreatedAt,
			&collab.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		collaborators = append(collaborators, collab)
	}

	return collaborators, nil
}

// GetUserPendingRequests - Gets all pending collaboration requests for a user
func (m *CollaboratorModel) GetUserPendingRequests(userID uuid.UUID) ([]Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Collaborator
	for rows.Next() {
		var collab Collaborator
		err := rows.Scan(
			&collab.ID,
			&collab.UserID,
			&collab.ProjectID,
			&collab.Status,
			&collab.CreatedAt,
			&collab.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, collab)
	}

	return requests, nil
}

// GetUserCollaborations - Gets all collaborations for a user (any status)
func (m *CollaboratorModel) GetUserCollaborations(userID uuid.UUID) ([]Collaborator, error) {
	query := `
		SELECT id, user_id, project_id, status, created_at, updated_at
		FROM collaborators
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborations []Collaborator
	for rows.Next() {
		var collab Collaborator
		err := rows.Scan(
			&collab.ID,
			&collab.UserID,
			&collab.ProjectID,
			&collab.Status,
			&collab.CreatedAt,
			&collab.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		collaborations = append(collaborations, collab)
	}

	return collaborations, nil
}

// DeleteCollaboration - Removes a collaboration
func (m *CollaboratorModel) DeleteCollaboration(collabID uuid.UUID) error {
	query := `DELETE FROM collaborators WHERE id = $1`
	_, err := m.DB.Exec(query, collabID)
	return err
}

// CheckCollaborationExists - Checks if a collaboration already exists
func (m *CollaboratorModel) CheckCollaborationExists(userID uuid.UUID, projectID uuid.UUID) (bool, error) {
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
