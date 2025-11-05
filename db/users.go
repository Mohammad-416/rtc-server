package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	GITHUB_ID  int64     `json:"github_id"`
	USERNAME   string    `json:"username"`
	EMAIL      string    `json:"email"`
	CREATED_AT time.Time `json:"created_at"`
}

type UserModel struct {
	DB *sql.DB
}

// Create a new user after GitHub sign-in
func (m *UserModel) CreateUser(githubID int64, username, email string) (User, error) {
	id := uuid.New()
	createdAt := time.Now()

	query := `
		INSERT INTO users (id, github_id, username, email, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.DB.Exec(query, id, githubID, username, email, createdAt)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:         id,
		GITHUB_ID:  githubID,
		USERNAME:   username,
		EMAIL:      email,
		CREATED_AT: createdAt,
	}, nil
}

func (m *UserModel) GetUser(username string) (User, error) {
	var user User
	query := `SELECT id, github_id, username, email, created_at FROM users WHERE username = $1`

	row := m.DB.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.GITHUB_ID, &user.USERNAME, &user.EMAIL, &user.CREATED_AT)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (m *UserModel) GetUserByID(userID uuid.UUID) (User, error) {
	var user User
	query := `SELECT id, github_id, username, email, created_at FROM users WHERE id = $1`

	row := m.DB.QueryRow(query, userID)
	err := row.Scan(&user.ID, &user.GITHUB_ID, &user.USERNAME, &user.EMAIL, &user.CREATED_AT)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (m *UserModel) GetUserByEmail(email string) (User, error) {
	var user User
	query := `SELECT id, github_id, username, email, created_at FROM users WHERE email = $1`

	row := m.DB.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.GITHUB_ID, &user.USERNAME, &user.EMAIL, &user.CREATED_AT)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (m *UserModel) GetAllUsers() ([]User, error) {
	query := `SELECT id, github_id, username, email, created_at FROM users`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.GITHUB_ID, &user.USERNAME, &user.EMAIL, &user.CREATED_AT)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (m *UserModel) UpdateUser(id uuid.UUID, username, email string) error {
	query := `
		UPDATE users SET username = $1, email = $2
		WHERE id = $3
	`

	_, err := m.DB.Exec(query, username, email, id)
	return err
}

func (m *UserModel) DeleteUser(username string) error {
	query := `DELETE FROM users WHERE username = $1`

	_, err := m.DB.Exec(query, username)
	return err
}
