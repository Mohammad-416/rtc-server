package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"` // References User.ID
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectModel struct {
	DB *sql.DB
}

func (m *ProjectModel) CreateProject(ownerID uuid.UUID, name, description string) (Project, error) {
	id := uuid.New()
	createdAt := time.Now()

	query := `
	INSERT INTO projects (id, owner_id, name, description, created_at)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.DB.Exec(query, id, ownerID, name, description, createdAt)
	if err != nil {
		return Project{}, err
	}

	return Project{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

func (m *ProjectModel) GetProjectsByUser(ownerID uuid.UUID) ([]Project, error) {
	query := `
	SELECT id, owner_id, name, description, created_at
	FROM projects
	WHERE owner_id = $1
	`

	rows, err := m.DB.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (m *ProjectModel) GetProjectByName(ownerID uuid.UUID, name string) (Project, error) {
	query := `
	SELECT id, owner_id, name, description, created_at
	FROM projects
	WHERE name = $1 AND owner_id = $2
	`

	var p Project
	err := m.DB.QueryRow(query, name, ownerID).Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt)
	if err != nil {
		return Project{}, err
	}

	return p, nil
}

func (m *ProjectModel) DeleteProject(ownerID uuid.UUID, projectID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1 AND owner_id = $2`
	_, err := m.DB.Exec(query, projectID, ownerID)
	return err
}
