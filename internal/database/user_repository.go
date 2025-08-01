package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/squ1ky/talkify/internal/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (ur *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := ur.db.QueryRow(query, user.Username, user.PasswordHash, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (ur *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE id = $1`

	err := ur.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (ur *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1`

	err := ur.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// List retrieves all users with pagination
func (ur *UserRepository) List(limit, offset int) ([]models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		ORDER by created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := ur.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// Count returns total number of users
func (ur *UserRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`

	err := ur.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Delete removes a user by ID
func (ur *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := ur.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Update updates user information (currently only username)
func (ur *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = $1 WHERE id = $2`

	result, err := ur.db.Exec(query, user.Username, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Exists checks if user with given username exists
func (ur *UserRepository) Exists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := ur.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}
