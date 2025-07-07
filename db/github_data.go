package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type GithubData struct {
	ID           uuid.UUID `json:"id"`
	GITHUB_TOKEN string    `json:"github_token"`
	USERNAME     string    `json:"username"`
	USER_ID      uuid.UUID `json:"user_id"`
	CREATED_AT   time.Time `json:"created_at"`
}

type TokenModel struct {
	DB *sql.DB
}

// Create a new user after GitHub sign-in
func (m *TokenModel) SaveToken(githubToken, username string, userID uuid.UUID) (GithubData, error) {
	id := uuid.New()
	createdAt := time.Now()

	query := `
		INSERT INTO github_data (id, github_token, username, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.DB.Exec(query, id, githubToken, username, userID, createdAt)
	if err != nil {
		return GithubData{}, err
	}

	return GithubData{
		ID:           id,
		GITHUB_TOKEN: githubToken,
		USERNAME:     username,
		USER_ID:      userID,
		CREATED_AT:   createdAt,
	}, nil
}

func (m *TokenModel) GetToken(username string) (GithubData, error) {
	var githubData GithubData
	query := `SELECT id, github_token, username, user_id, created_at FROM github_data WHERE username = $1`

	row := m.DB.QueryRow(query, username)
	err := row.Scan(&githubData.ID, &githubData.GITHUB_TOKEN, &githubData.USERNAME, &githubData.USER_ID, &githubData.CREATED_AT)
	if err != nil {
		return GithubData{}, err
	}
	return githubData, nil
}

func (m *TokenModel) UpdateToken(username, githubToken string) error {
	query := `
		UPDATE github_data SET github_token = $1
		WHERE username = $2
	`

	_, err := m.DB.Exec(query, githubToken, username)
	return err
}
